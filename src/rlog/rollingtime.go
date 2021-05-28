package rlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type rollingTimeFile struct {
	sync.Mutex

	file           *os.File
	filePath       string
	fileCreateTime time.Time
	closed         bool

	basePath   string
	jobs       *Jobs
	timeLayout string
}

func NewRollingTimeFile(basePath, crontabStr string) (io.WriteCloser, error) {
	basePath = strings.TrimRight(basePath, ".log")
	jobs := NewJobList()
	rollFile := new(rollingTimeFile)
	rollFile.timeLayout = "20060102-1504"
	rollFile.basePath = basePath
	rollFile.jobs = jobs
	rollFile.filePath = fmt.Sprintf("%s.log", rollFile.basePath)
	err := jobs.AddJob(crontabStr, func() {
		rollFile.Lock()
		defer rollFile.Unlock()
		rollFile.roll()
	})
	if err != nil {
		return nil, err
	}
	// 先触发一次
	rollFile.roll()
	// 启动定时器
	jobs.Process()
	return rollFile, nil
}

// roll 应该由定时器调用
func (r *rollingTimeFile) roll() error {
	// 需要切换创建文件
	if r.file != nil {
		r.file.Close()
		createTime := fmt.Sprintf("%s.%s.log", r.basePath, r.fileCreateTime.Format(r.timeLayout))
		os.Rename(r.filePath, createTime)
		r.file = nil
	}
	// 生成新文件
	if r.file == nil {
		f, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return TraceErr(err)
		} else {
			r.file = f
			r.fileCreateTime = time.Now()
			return nil
		}
	}
	return nil
}

func (r *rollingTimeFile) Close() error {
	r.Lock()
	defer r.Unlock()
	if r.closed {
		return nil
	}

	log.Printf("关闭")

	r.closed = true
	r.jobs.Close()
	if f := r.file; f != nil {
		r.file = nil
		return TraceErr(f.Close())
	}
	return nil
}

func (r *rollingTimeFile) Write(b []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	if r.closed {
		return 0, TraceErr(ErrClosedRollingFile)
	}

	// 写入内存
	n, err := r.file.Write(b)
	// r.file.Sync() 同步到磁盘
	if err != nil {
		return n, TraceErr(err)
	} else {
		return n, nil
	}
}
