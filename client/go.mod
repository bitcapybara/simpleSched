module github.com/bitcapybara/simpleSched/client

go 1.16

require (
	github.com/bitcapybara/cuckoo/client v0.0.1
	github.com/bitcapybara/cuckoo/core v0.0.1
	github.com/bitcapybara/raft v0.0.1
)

replace (
	github.com/bitcapybara/cuckoo/client => ../../cuckoo/client
	github.com/bitcapybara/cuckoo/core => ../../cuckoo/core
	github.com/bitcapybara/raft => ../../raft
)