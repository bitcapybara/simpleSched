module github.com/bitcapybara/simpleSched/server

go 1.16

require (
	github.com/bitcapybara/cuckoo/core v0.0.1
	github.com/bitcapybara/cuckoo/server v0.0.1
	github.com/bitcapybara/raft v0.0.1
	github.com/gin-gonic/gin v1.7.1
	github.com/go-playground/validator/v10 v10.5.0 // indirect
	github.com/go-resty/resty/v2 v2.6.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/ugorji/go v1.2.5 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781 // indirect
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/bitcapybara/cuckoo/core => ../../cuckoo/core
	github.com/bitcapybara/cuckoo/server => ../../cuckoo/server
	github.com/bitcapybara/raft => ../../raft
)
