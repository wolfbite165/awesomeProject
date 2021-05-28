// +build freebsd openbsd netbsd dragonfly darwin linux

package rlog

import (
	"os"
	"runtime"
	"syscall"
)

func HandleStdLog(path string) {
	Infof("operating system: %s", runtime.GOOS)
	// 将stdout, stderr输出到文件
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	err = syscall.Dup2(int(f.Fd()), 2)
	if err != nil {
		panic(err)
	}
	err = syscall.Dup2(int(f.Fd()), 1)
	if err != nil {
		panic(err)
	}
}
