package main

import (
	"flag"
	"fmt"
	"github.com/bitcapybara/cuckoo/core"
	"github.com/bitcapybara/raft"
	"github.com/bitcapybara/simpleSched/client/job"
	"github.com/bitcapybara/simpleSched/client/raftimpl"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"time"
)

func main() {

	var localAddr string
	flag.StringVar(&localAddr, "local", "", "当前节点通信地址")

	var remoteAddr string
	flag.StringVar(&remoteAddr, "remote", "", "服务端通信地址")

	flag.Parse()

	if localAddr == "" {
		log.Fatal("未指定当前节点通信地址")
	}

	if remoteAddr == "" {
		log.Fatal("未指定服务端通信地址")
	}

	newSchedClient(localAddr, remoteAddr).start()
}

type schedClient struct {
	localAddr  core.NodeAddr
	remoteAddr core.NodeAddr
	ginServer  *gin.Engine
	logger     raft.Logger
}

func newSchedClient(local, remote string) *schedClient {
	ginServer := gin.Default()
	return &schedClient{
		localAddr:  core.NodeAddr(local),
		remoteAddr: core.NodeAddr(remote),
		ginServer:  ginServer,
		logger:     raftimpl.NewLogger(),
	}
}

func (c *schedClient) start() {
	// 定期发送心跳
	go c.heartbeat()
	// 提交任务
	go c.submitJobs()
	// 创建路由
	router := c.ginServer
	router.POST("/execute", c.execute)

	// 启动服务
	_ = router.Run(string(c.localAddr))
}

func (c *schedClient) heartbeat() {
	tick := time.Tick(time.Second * 3)
	client := resty.New()
	req := core.HeartbeatReq{
		Group:     "simpleSched",
		LocalAddr: c.localAddr,
	}
	// 启动时先发一次心跳
	c.sendHeartbeat(client, req)
	for {
		<-tick
		c.sendHeartbeat(client, req)
	}
}

func (c *schedClient) sendHeartbeat(client *resty.Client, req core.HeartbeatReq) {
	// 发送请求
	url := fmt.Sprintf("http://%s/heartbeat", c.remoteAddr)
	var res core.HeartbeatReply
	response, resErr := client.R().SetHeader("Content-Type", "application/json").SetBody(req).SetResult(&res).Post(url)
	if resErr != nil {
		c.logger.Error(fmt.Errorf("发送请求失败！%w", resErr).Error())
	}
	if response.StatusCode() != 200 {
		c.logger.Error(fmt.Errorf("发送请求响应码异常：%d", response.StatusCode()).Error())
	}
}

func (c *schedClient) submitJobs() {
	jobs := job.LoadJobs(string(c.localAddr))
	client := resty.New()
	for _, submitJob := range jobs {
		//if submitJob.Id == "cron" {
		//	submitJob.Enable = true
		//}
		submitJob.Enable = true
		req := core.AddJobReq{
			Job:    submitJob,
		}
		// 发送请求
		url := fmt.Sprintf("http://%s/job/add", c.remoteAddr)
		var res core.CudReply
		response, resErr := client.R().SetHeader("Content-Type", "application/json").SetBody(req).SetResult(&res).Post(url)
		if resErr != nil {
			c.logger.Error(fmt.Errorf("发送请求失败！%w", resErr).Error())
		}
		if response.StatusCode() != 200 {
			c.logger.Error(fmt.Errorf("发送请求响应码异常：%d", response.StatusCode()).Error())
		}
		if res.Status != core.Ok {
			c.logger.Error("添加任务失败！")
		}
		c.logger.Trace(fmt.Sprintf("任务 id=%s 发送成功", submitJob.Id))
	}
}

func (c *schedClient) execute(ctx *gin.Context) {
	jobId := ctx.Query("jobId")
	c.logger.Trace(fmt.Sprintf("接收到 id=%s 任务执行请求", jobId))
	err := job.ExecuteJob(core.JobId(jobId))
	if err != nil {
		ctx.String(500, "执行任务出错: "+err.Error())
		return
	}
	ctx.String(http.StatusOK, "OK")
}
