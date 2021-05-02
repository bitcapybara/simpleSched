package main

import (
	"flag"
	"fmt"
	"github.com/bitcapybara/cuckoo/core"
	"github.com/bitcapybara/raft"
	"github.com/bitcapybara/simpleSched/client/job"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"net/http"
	"time"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "当前节点nodeId")

	var localAddr string
	flag.StringVar(&port, "local", "localhost:8080", "当前节点通信地址")

	var remoteAddr string
	flag.StringVar(&port, "remote", "localhost:8080", "服务端通信地址")

	newSchedClient(localAddr, remoteAddr).start(port)
}

type schedClient struct {
	localAddr  core.NodeAddr
	remoteAddr core.NodeAddr
	ginServer  *gin.Engine

	logger raft.Logger
}

func newSchedClient(local, remote string) *schedClient {
	ginServer := gin.Default()
	return &schedClient{
		localAddr:  core.NodeAddr(local),
		remoteAddr: core.NodeAddr(remote),
		ginServer:  ginServer,
	}
}

func (c *schedClient) start(port string) {
	// 定期发送心跳
	go c.heartbeat()
	// 提交任务
	go c.submitJobs()
	// 创建路由
	router := c.ginServer
	router.POST("/execute", c.execute)

	// 启动服务
	_ = router.Run(":" + port)
}

func (c *schedClient) heartbeat() {
	tick := time.Tick(time.Second * 3)
	client := resty.New()
	req := core.HeartbeatReq{
		Group:     "test",
		LocalAddr: c.localAddr,
	}
	for {
		<-tick
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
}

func (c *schedClient) submitJobs() {
	jobs := job.LoadJobs(string(c.localAddr))
	client := resty.New()
	for _, submitJob := range jobs {
		req := core.AddJobReq{
			Job: submitJob,
			Enable: true,
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
	}
}

func (c *schedClient) execute(ctx *gin.Context) {
	jobId := ctx.Param("jobId")
	err := job.ExecuteJob(core.JobId(jobId))
	if err != nil {
		ctx.String(500, "执行任务出错: "+err.Error())
	}
	ctx.String(http.StatusOK, "OK")
}
