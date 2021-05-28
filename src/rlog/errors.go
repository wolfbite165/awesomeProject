package rlog

import (
	"fmt"
)

// TracedError 封装一个被跟踪的 err
type TracedError struct {
	// Stack 跟踪的n层调用栈
	Stack Stack
	// Cause 针对此err进行的跟踪
	Cause error
	// 附加信息
	Extra string
}

// Error 实现了 error 接口
func (e *TracedError) Error() string {
	return e.Cause.Error()
}

func TraceErr(err error) error {
	// 如果不跟踪或者err为nil
	if err == nil {
		return err
	}
	// 如果已经跟踪过
	_, ok := err.(*TracedError)
	if ok {
		return err
	}
	// 对原生的 error 进行跟踪.
	return &TracedError{
		Stack: TraceN(1, 32), //跳过TraceErr函数, 向上32层
		Cause: err,
	}
}

func TraceErrf(err error, format string, v ...interface{}) error {
	// 如果不跟踪或者err为nil
	if err == nil {
		return err
	}
	// 如果已经跟踪过
	_, ok := err.(*TracedError)
	if ok {
		return err
	}
	// 对原生的 error 进行跟踪.
	return &TracedError{
		Stack: TraceN(1, 32), //跳过TraceErr函数, 向上32层
		Cause: err,
		Extra: fmt.Sprintf("\n"+format, v...),
	}
}

// NewErrorf 生成一个被跟踪的 error
func NewErrorf(format string, v ...interface{}) error {
	err := fmt.Errorf(format, v...)
	return &TracedError{
		Stack: TraceN(1, 32),
		Cause: err,
	}
}

// GetStack 获取一个被跟踪的 error 的 stack 字段
func GetStack(err error) Stack {
	if err == nil {
		return nil
	}
	e, ok := err.(*TracedError)
	if ok {
		return e.Stack
	}
	return nil
}

func GetCause(err error) error {
	for err != nil {
		e, ok := err.(*TracedError)
		if ok {
			err = e.Cause
		} else {
			return err
		}
	}
	return nil
}

func Equal(err1, err2 error) bool {
	e1 := GetCause(err1)
	e2 := GetCause(err2)
	if e1 == e2 {
		return true
	}
	if e1 == nil || e2 == nil {
		return e1 == e2
	}
	return e1.Error() == e2.Error()
}

func NotEqual(err1, err2 error) bool {
	return !Equal(err1, err2)
}
