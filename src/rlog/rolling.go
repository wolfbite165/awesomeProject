package rlog

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
)

type rollingFile struct {
	mu sync.Mutex

	closed bool

	maxFileFrag int
	maxFragSize int64

	file     *os.File
	basePath string
	filePath string
	fileFrag int
	fragSize int64
}

var ErrClosedRollingFile = NewErrorf("rolling file is closed")

func (r *rollingFile) roll() error {
	if r.file != nil {
		if r.fragSize < r.maxFragSize {
			return nil
		}
		r.file.Close()
		r.file = nil
	}
	r.fragSize = 0
	r.fileFrag = (r.fileFrag + 1) % r.maxFileFrag
	r.filePath = fmt.Sprintf("%s.%d", r.basePath, r.fileFrag)

	f, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return TraceErr(err)
	} else {
		r.file = f
		return nil
	}
}

func (r *rollingFile) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	if f := r.file; f != nil {
		r.file = nil
		return TraceErr(f.Close())
	}
	return nil
}

func (r *rollingFile) Write(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return 0, TraceErr(ErrClosedRollingFile)
	}

	if err := r.roll(); err != nil {
		return 0, err
	}

	// 写入内存
	n, err := r.file.Write(b)
	// r.file.Sync() 同步到磁盘
	r.fragSize += int64(n)
	if err != nil {
		return n, TraceErr(err)
	} else {
		return n, nil
	}
}

func NewRollingFile(basePath string, maxFileFrag int, maxFragSize int64) (io.WriteCloser, error) {
	if maxFileFrag <= 0 {
		return nil, NewErrorf("invalid max file-frag = %d", maxFileFrag)
	}
	if maxFragSize <= 0 {
		return nil, NewErrorf("invalid max frag-size = %d", maxFragSize)
	}
	if _, file := path.Split(basePath); file == "" {
		return nil, NewErrorf("invalid base-path = %s, file name is required", basePath)
	}

	var fileFrag = 0
	for i := 0; i < maxFileFrag; i++ {
		_, err := os.Stat(fmt.Sprintf("%s.%d", basePath, i))
		if err != nil && os.IsNotExist(err) {
			fileFrag = i
			break
		}
	}

	return &rollingFile{
		maxFileFrag: maxFileFrag,
		maxFragSize: maxFragSize,

		basePath: basePath,
		fileFrag: fileFrag - 1,
	}, nil
}
