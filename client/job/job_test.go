package job

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestJob(t *testing.T) {
	fName := fmt.Sprintf("/Users/yuxingy/Downloads/cron-%s", time.Now().Format("2006-01-02_15-04-05.000"))
	f, err := os.Create(fName)
	if err != nil {
		t.Error(fmt.Errorf("创建文件失败！%w", err))
	}
	f.Close()
}
