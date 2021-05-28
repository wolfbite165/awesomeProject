package rlog

import (
	"log"
	"os"
	"reflect"
	"testing"
)

// TAssert fails the test if the condition is false.
func TAssert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		Errorf(msg, v...)
		tb.FailNow()
	}
}

// TOk fails the test if an err is not nil.
func TOk(tb testing.TB, err error) {
	if err != nil {
		Errorf("unexpected error: %s", err.Error())
		tb.FailNow()
	}
}

// TEquals fails the test if exp is not equal to act.
func TEquals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		Errorf("exp: %#v\n got: %#v", exp, act)
		tb.FailNow()
	}
}

// testLoggerAdapter 将 testing 变为输出流
type testLoggerAdapter struct {
	tb testing.TB
}

// Write 实现 io.Writer 接口, 作为 log.New的第一个参数
func (a *testLoggerAdapter) Write(d []byte) (int, error) {
	// testing 自动会加上换行, 所以这里删除换行
	if d[len(d)-1] == '\n' {
		d = d[:len(d)-1]
	}
	a.tb.Log(string(d))
	return len(d), nil
}

// Help_SetLogOutput 使用测试的输出 (没什么毛用)
func Help_SetLogOutput(tb testing.TB) {
	ad := &testLoggerAdapter{tb}
	log.SetOutput(ad)
	StdLog = New(ad, "")
}

// Help_NewTestLogger 创建一个测试适配器 (没什么毛用)
func Help_NewTestLogger(tb testing.TB) *log.Logger {
	if tb == nil {
		return log.New(os.Stderr, "", 0)
	}
	return log.New(&testLoggerAdapter{tb}, "", 0)
}
