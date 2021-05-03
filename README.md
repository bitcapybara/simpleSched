# simpleSched

### 客户端（client包）

* 实现了两个简单任务，`cron` 和 `fixedDelay` ，分别在指定目录中新建文件
* 定期向服务端发送心跳
* 接收来自服务端的任务执行请求，执行相应任务

### 服务端（server包）

* `/job/add`：接收客户端添加任务的请求
* `/heartbeat`：接收客户端的心跳请求
* `/appendEntries`，`/requestVote`，`/installSnapshot`：接收来自其它节点的 raft 请求