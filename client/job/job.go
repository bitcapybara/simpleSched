package job

import (
	"fmt"
	"github.com/bitcapybara/cuckoo/core"
	"os"
	"time"
)

type job interface {

	genJob() core.Job

	execute() error
}

type cronJob struct {

}

func newCronJob() cronJob {
	return cronJob{}
}

func (c cronJob) genJob() core.Job {
	return core.Job{
		Id: "cron",
		ScheduleRule: core.ScheduleRule{
			ScheduleType: core.Cron,
			CronExpr: "0/5 * * * *",
			ParseOption: core.Second | core.Minute | core.Hour | core.Dom | core.Month,
		},
		Router: core.First,
	}
}

func (c cronJob) execute() error {
	fName := fmt.Sprintf("/Users/yuxingy/Downloads/cron-%s", time.Now().Format("2006-01-02_15-04-05.000"))
	_, err := os.Create(fName)
	if err != nil {
		return fmt.Errorf("创建文件失败！%w", err)
	}
	return nil
}

type fixRateJob struct {

}

func newFixRateJob() fixRateJob {
	return fixRateJob{}
}

func (c fixRateJob) genJob() core.Job {
	return core.Job{
		Id: "fixDelay",
		ScheduleRule: core.ScheduleRule{
			ScheduleType: core.FixedRate,
			Initial: time.Second * 5,
			Duration: time.Second * 3,
		},
		Router: core.First,
	}
}

func (c fixRateJob) execute() error {
	fName := fmt.Sprintf("/Users/yuxingy/Downloads/fixedRate-%s", time.Now().Format("2006-01-02_15-04-05.000"))
	_, err := os.Create(fName)
	if err != nil {
		return fmt.Errorf("创建文件失败！%w", err)
	}
	return nil
}
