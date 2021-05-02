package job

import (
	"errors"
	"github.com/bitcapybara/cuckoo/core"
)

var jobMap map[core.JobId]job

func init() {
	cron := newCronJob()
	delay := newFixDelayJob()
	jobMap = map[core.JobId]job{
		cron.genJob().Id: cron,
		delay.genJob().Id: delay,
	}
}

func LoadJobs(path string) []core.Job {
	result := make([]core.Job, 0)
	for _, jb := range jobMap {
		genJob := jb.genJob()
		genJob.Path = path
		genJob.Group = "simpleSched"
		result = append(result, genJob)
	}
	return result
}

func ExecuteJob(id core.JobId) error {
	if j, ok := jobMap[id]; ok {
		return j.execute()
	}
	return errors.New("任务不存在！")
}


