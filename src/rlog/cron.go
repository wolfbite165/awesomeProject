package rlog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func compare(current int64, scheduled string) bool {
	// 通配
	if scheduled == "*" {
		return true
	}

	// 试图按照10进制和64bit转成数值, 如果相等则说明匹配上
	if numeric, err := strconv.ParseInt(scheduled, 10, 64); err == nil {
		return numeric == current
	}

	// 如果开头是 '*/' 说明后面是整数, 作为分母
	if scheduled[:2] == "*/" {
		divider, err := strconv.ParseInt(scheduled[2:], 10, 64)
		if err != nil || divider == 0 {
			return false
		}
		return current%divider == 0
	}

	return false
}

type Job struct {
	min       string
	hour      string
	dom       string
	mon       string
	dow       string
	task      func()
	lastCheck time.Time
}

func (job *Job) String() string {
	return fmt.Sprintf("min=%s hour=%s dom=%s mon=%s dow=%s last=%v",
		job.min,
		job.hour,
		job.dom,
		job.mon,
		job.dow,
		job.lastCheck,
	)
}

func (job *Job) IsMatchTime(t time.Time) bool {
	// 避免在一分钟内, 重复执行, 因为此函数每秒钟调用一次
	if job.lastCheck.Unix()/60 >= t.Unix()/60 {
		return false
	}
	job.lastCheck = t

	// 上面确保每次进来都过了一分钟
	return compare(int64(t.Minute()), job.min) &&
		compare(int64(t.Hour()), job.hour) &&
		compare(int64(t.Day()), job.dom) &&
		compare(int64(t.Month()), job.mon) &&
		compare(int64(t.Weekday()), job.dow)
}

type Jobs struct {
	sync.Mutex
	jobSlice   []*Job
	started    bool
	reSchedule *regexp.Regexp
	closed     chan struct{}
}

func NewJobList() *Jobs {
	jobs := new(Jobs)
	jobs.started = false
	jobs.reSchedule = regexp.MustCompile(`(\*\/[0-9]{1,2})|(\*)|([0-9]{1,2})`)
	jobs.jobSlice = make([]*Job, 0, 10)
	jobs.closed = make(chan struct{}, 1)
	return jobs
}

func (jobs *Jobs) AddJob(schedule string, task func()) error {
	matches := jobs.reSchedule.FindAllString(schedule, -1)
	if len(matches) != 5 {
		return errors.New(`Schedule should be specified as %min %hour %dom %mon %dow`)
	}

	job := &Job{
		min:       matches[0],
		hour:      matches[1],
		dom:       matches[2],
		mon:       matches[3],
		dow:       matches[4],
		lastCheck: time.Now().AddDate(-200, 0, 0),
		task:      task,
	}

	jobs.Lock()
	jobs.jobSlice = append(jobs.jobSlice, job)
	jobs.Unlock()
	return nil
}

func (jobs *Jobs) Close() {
	select {
	case jobs.closed <- struct{}{}:
	default:
	}
}

// Process 执行定时任务
func (jobs *Jobs) Process() {
	jobs.Lock()
	if jobs.started {
		jobs.Unlock()
		return
	}
	jobs.started = true
	jobs.Unlock()

	go func() {
		// 每N秒钟检查一次, crontab 的最小刻度是分钟, 所以没什么问题
		ticker := time.Tick(time.Second * 10)
		for {
			select {
			case <-jobs.closed:
				break
			case <-ticker:
				now := time.Now()
				for _, job := range jobs.jobSlice {
					if job.IsMatchTime(now) {
						go job.task()
					}
				}

			}
		}
	}()
}

// 按照前缀保留一段时间的文件:
//  例如:　KeepFilePrefix（"./a/b/ss.cc.dd.",time.Hour * 24 * 15）
//  删除 15 天以前的文件( ./a/b/ss.cc.dd.sdfd,  ./a/b/ss.cc.dd.cc, ... )
func KeepFilePrefix(prefix string, dur time.Duration) {
	go func() {
		dir := filepath.Dir(prefix)
		for {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if strings.HasPrefix(path, prefix) {
					if info.ModTime().Before(time.Now().Add(-dur)) {
						Infof("delete %s", path)
						err := os.Remove(path)
						ErrorError(err)
					}
				}
				return nil
			})
			time.Sleep(time.Minute)
		}
	}()
}
