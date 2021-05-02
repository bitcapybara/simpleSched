package cuckooimpl

import (
	"fmt"
	"github.com/bitcapybara/cuckoo/core"
	"github.com/bitcapybara/raft"
	"github.com/go-resty/resty/v2"
)

type HttpDispatcher struct {
	logger raft.Logger
	client *resty.Client
}

func NewDispatcher(logger raft.Logger) *HttpDispatcher {
	return &HttpDispatcher{
		logger: logger,
		client: resty.New(),
	}
}

func (h *HttpDispatcher) Dispatch(clientAddr core.NodeAddr, id core.JobId) (err error) {
	defer func() {
		if err != nil {
			h.logger.Error(err.Error())
		}
	}()
	// 发送请求
	url := fmt.Sprintf("%s%s%s", "http://", clientAddr, "/execute")
	var res string
	response, resErr := h.client.R().SetQueryParam("jobId", string(id)).SetResult(&res).Post(url)
	if resErr != nil {
		return fmt.Errorf("发送请求失败！%w, %s", resErr)
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("发送请求响应码异常：%d, %s", response.StatusCode())
	}
	return nil
}



