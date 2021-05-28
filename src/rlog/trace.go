package rlog

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

// 定义tab
const tab = "    "

// Record  一条跟踪记录
type Record struct {
	// Name 记录名字
	Name string
	// File 记录在的文件
	File string
	// Line 记录在的行数
	Line int
}

// String 字符串化方法
func (r *Record) String() string {
	if r == nil {
		return "[nil-record]"
	}
	return fmt.Sprintf("%s:%d %s", r.File, r.Line, r.Name)
}

// Stack 一个调用栈
type Stack []*Record

func Trace() Stack {
	// 第一个参数1, 代表跳过 Trace 函数.
	return TraceN(1, 32)
}

func (s Stack) String() string {
	return s.StringWithIndent(0)
}

func (s Stack) StringWithIndent(indent int) string {
	var b bytes.Buffer
	for i, r := range s {
		for j := 0; j < indent; j++ {
			fmt.Fprint(&b, tab)
		}
		fmt.Fprintf(&b, "%-3d %s:%d\n", len(s)-i-1, r.File, r.Line)
		for j := 0; j < indent; j++ {
			fmt.Fprint(&b, tab)
		}
		fmt.Fprint(&b, tab, tab)
		fmt.Fprint(&b, r.Name, "\n")
	}
	if len(s) != 0 {
		for j := 0; j < indent; j++ {
			fmt.Fprint(&b, tab)
		}
		fmt.Fprint(&b, tab, "... ...\n")
	}
	return b.String()
}

// TraceN 跳过 skip 层
// 向高层跟踪 depth 层
func TraceN(skip, depth int) Stack {
	skip += 1 //跳过跟踪当前 TraceN 函数
	s := make([]*Record, 0, depth)
	for i := 0; i < depth; i++ {
		r := Caller(skip + i)
		if r == nil {
			break
		}
		s = append(s, r)
	}
	return s
}

// Caller 找到调用者, skip >= 0 才有意义.
// skip = 0: 返回调用 Caller 的函数
// skip = -1: 返回调用 Caller 的函数内部调用的 runtimeCaller
// skip = 1: 返回调用 Caller 的函数的调用者
func Caller(skip int) *Record {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return nil
	}
	// 如果跟踪到 runtime 则返回nil
	fn := runtime.FuncForPC(pc)
	if fn == nil || strings.HasPrefix(fn.Name(), "runtime.") {
		return nil
	}
	return &Record{
		Name: fn.Name(),
		File: file,
		Line: line,
	}
}
