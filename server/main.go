package main

import (
	"flag"
	"github.com/bitcapybara/cuckoo/core"
	"github.com/bitcapybara/cuckoo/server"
	"github.com/bitcapybara/cuckoo/server/controller"
	"github.com/bitcapybara/raft"
	"github.com/bitcapybara/simpleSched/server/cuckooimpl"
	"github.com/bitcapybara/simpleSched/server/raftimpl"
	"github.com/gin-gonic/gin"
	"log"
	"strings"
	"time"
)

const NoneOption = ""

func main() {
	// 命令行参数定义
	var me string
	flag.StringVar(&me, "me", NoneOption, "当前节点nodeId")
	var peerStr string
	flag.StringVar(&peerStr, "peers", NoneOption, "指定所有节点地址，nodeId@nodeAddr，多个地址使用逗号间隔")
	var role string
	flag.StringVar(&role, "role", NoneOption, "当前节点角色")
	flag.Parse()

	// 命令行参数解析
	if me == "" {
		log.Fatal("未指定当前节点id！")
	}

	if peerStr == "" {
		log.Fatalln("未指定集群节点")
	}

	if role == "" {
		log.Fatalln("未指定节点角色")
	}

	peerSplit := strings.Split(peerStr, ",")
	peers := make(map[raft.NodeId]raft.NodeAddr, len(peerSplit))
	for _, peerInfo := range peerSplit {
		idAndAddr := strings.Split(peerInfo, "@")
		peers[raft.NodeId(idAndAddr[0])] = raft.NodeAddr(idAndAddr[1])
	}

	s := newServer(raft.RoleFromString(role), raft.NodeId(me), peers)
	s.start()
}

type schedServer struct {
	addr         string
	cuckooServer *server.Server
	ginServer    *gin.Engine
	logger       raft.Logger
}

func newServer(role raft.RoleStage, me raft.NodeId, peers map[raft.NodeId]raft.NodeAddr) *schedServer {
	logger := raftimpl.NewLogger()
	config := server.Config{
		RaftConfig: raft.Config{
			Peers:              peers,
			Me:                 me,
			Role:               role,
			Transport:          raftimpl.NewHttpTransport(logger),
			Logger:             logger,
			RaftStatePersister: raftimpl.NewRaftStatePersister(),
			SnapshotPersister:  raftimpl.NewSnapshotPersister(),
			ElectionMaxTimeout: 10000,
			ElectionMinTimeout: 5000,
			HeartbeatTimeout:   1000,
			MaxLogLength:       50,
		},
		JobPool:        controller.NewSliceJobPool(logger),
		JobDispatcher:  cuckooimpl.NewDispatcher(logger),
		ExecutorExpire: time.Second * 10,
	}
	ginServer := gin.Default()
	return &schedServer{
		addr:         string(peers[me]),
		cuckooServer: server.NewServer(config),
		ginServer:    ginServer,
		logger:       logger,
	}
}

func (s *schedServer) start() {
	// 启动 cuckoo 服务
	go s.cuckooServer.Start()

	// 启动 gin服务
	g := s.ginServer
	g.POST("/job/add", s.addJob)
	g.POST("/heartbeat", s.heartbeat)

	g.POST("/appendEntries", s.appendEntries)
	g.POST("/requestVote", s.requestVote)
	g.POST("/installSnapshot", s.installSnapshot)
	_ = g.Run(s.addr)
}

func (s *schedServer) addJob(ctx *gin.Context) {
	// 反序列化获取请求参数
	var args core.AddJobReq
	bindErr := ctx.Bind(&args)
	if bindErr != nil {
		ctx.String(500, "反序列化参数失, %s", bindErr.Error())
		return
	}
	var res core.CudReply
	cuckooErr := s.cuckooServer.AddJob(args, &res)
	if cuckooErr != nil {
		ctx.String(500, "cuckoo 操作失败！%s", cuckooErr.Error())
		return
	}
	// 序列化并返回结果
	ctx.JSON(200, res)
}

func (s *schedServer) heartbeat(ctx *gin.Context) {
	// 反序列化获取请求参数
	var args core.HeartbeatReq
	bindErr := ctx.Bind(&args)
	if bindErr != nil {
		ctx.String(500, "反序列化参数失, %s", bindErr.Error())
		return
	}
	var res core.HeartbeatReply
	cuckooErr := s.cuckooServer.Heartbeat(args, &res)
	if cuckooErr != nil {
		ctx.String(500, "cuckoo 操作失败！%s", cuckooErr.Error())
		return
	}
	// 序列化并返回结果
	ctx.JSON(200, res)
}

func (s *schedServer) appendEntries(ctx *gin.Context) {
	// 反序列化获取请求参数
	var args raft.AppendEntry
	bindErr := ctx.Bind(&args)
	if bindErr != nil {
		ctx.String(500, "反序列化参数失, %s", bindErr.Error())
		return
	}
	// 调用 raft 逻辑
	var res raft.AppendEntryReply
	raftErr := s.cuckooServer.AppendEntries(args, &res)
	if raftErr != nil {
		ctx.String(500, "raft 操作失败！%s", raftErr.Error())
		return
	}
	// 序列化并返回结果
	ctx.JSON(200, res)
}

func (s *schedServer) requestVote(ctx *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			s.logger.Error(err.Error())
		}
	}()
	// 反序列化获取请求参数
	var args raft.RequestVote
	bindErr := ctx.Bind(&args)
	if bindErr != nil {
		ctx.String(500, "反序列化参数失败！%s", bindErr.Error())
		return
	}
	// 调用 raft 逻辑
	var res raft.RequestVoteReply
	raftErr := s.cuckooServer.RequestVote(args, &res)
	if raftErr != nil {
		ctx.String(500, "raft 操作失败！%s", raftErr.Error())
		return
	}
	// 序列化并返回结果
	ctx.JSON(200, res)
}

func (s *schedServer) installSnapshot(ctx *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			s.logger.Error(err.Error())
		}
	}()
	// 反序列化获取请求参数
	var args raft.InstallSnapshot
	bindErr := ctx.Bind(&args)
	if bindErr != nil {
		ctx.String(500, "反序列化参数失败！%s", bindErr.Error())
		return
	}
	// 调用 raft 逻辑
	var res raft.InstallSnapshotReply
	raftErr := s.cuckooServer.InstallSnapshot(args, &res)
	if raftErr != nil {
		ctx.String(500, "raft 操作失败！%s", raftErr.Error())
		return
	}
	// 序列化并返回结果
	ctx.JSON(200, res)
}
