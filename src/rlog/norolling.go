package rlog

import (
	"io"
	"os"
	"path"
	"sync"
)

type noRollingFile struct {
	mu       sync.Mutex
	closed   bool
	file     *os.File
	basePath string
}

var ErrClosedNoRollingFile = NewErrorf("noRolling file is closed")

func (r *noRollingFile) checkFileExist() error {
	// 文件不存在时, 确保文件引用为nil
	if _, err := os.Stat(r.basePath); os.IsNotExist(err) {
		if r.file != nil {
			r.file.Close()
			r.file = nil
		}
	}
	// 如果文件引用为nil, 需要打开或者创建文件.
	if r.file == nil {
		f, err := os.OpenFile(r.basePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return TraceErr(err)
		} else {
			r.file = f
		}
	}
	return nil
}

func (r *noRollingFile) Close() error {
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

func (r *noRollingFile) Write(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return 0, TraceErr(ErrClosedNoRollingFile)
	}

	if err := r.checkFileExist(); err != nil {
		return 0, err
	}

	n, err := r.file.Write(b)
	if err != nil {
		return n, TraceErr(err)
	} else {
		return n, nil
	}
}

func NewNoRollingFile(basePath string) (io.WriteCloser, error) {
	if _, file := path.Split(basePath); file == "" {
		return nil, NewErrorf("invalid base-path = %s, file name is required", basePath)
	}
	return &noRollingFile{
		basePath: basePath,
	}, nil
}
