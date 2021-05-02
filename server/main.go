package main

import (
	"flag"
	"github.com/bitcapybara/cuckoo/server"
	"github.com/bitcapybara/cuckoo/server/controller"
	"github.com/bitcapybara/raft"
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
}

func newServer(role raft.RoleStage, me raft.NodeId, peers map[raft.NodeId]raft.NodeAddr) *schedServer {
	logger := raftimpl.NewLogger()
	config := server.Config{
		RaftConfig: raft.Config{
			Peers:              peers,
			Me:                 me,
			Role:               role,
			ElectionMaxTimeout: 10000,
			ElectionMinTimeout: 5000,
			HeartbeatTimeout:   1000,
			MaxLogLength:       50,
		},
		JobPool:       controller.NewSliceJobPool(logger),
		JobDispatcher: nil,
		ExecutorExpire: time.Second * 10,
	}
	ginServer := gin.Default()
	return &schedServer{
		addr:   string(peers[me]),
		cuckooServer: server.NewServer(config),
		ginServer:    ginServer,
	}
}

func (s *schedServer) start() {
	// 启动 cuckoo 服务
	go s.cuckooServer.Start()

	// 启动 gin服务
	g := s.ginServer
	g.GET("/job/add")
	_ = g.Run(s.addr)
}
