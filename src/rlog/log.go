package rlog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type Placehoder struct {
	Picker string `json:"picker"`
}

const (
	Ldate         = log.Ldate
	Llongfile     = log.Llongfile
	Lmicroseconds = log.Lmicroseconds
	Lshortfile    = log.Lshortfile
	LstdFlags     = log.LstdFlags
	Ltime         = log.Ltime
)
const (
	TYPE_ERROR = LogType(1 << iota)
	TYPE_WARN
	TYPE_INFO
	TYPE_DEBUG
	TYPE_PANIC = LogType(^0)
)

const (
	LEVEL_NONE = LogLevel(1<<iota - 1)
	LEVEL_ERROR
	LEVEL_WARN
	LEVEL_INFO
	LEVEL_DEBUG
)

const MAYBE_INCONSISTENCY = "maybe_inconsistency"

type (
	// LogType 表示 log 的类型, 如 info, debug
	LogType int64
	// LogLevel 表示 log 的级别, 如打印 info 以上的log, 还是打印debug以上的log.
	LogLevel int64
)

func (t LogType) String() string {
	switch t {
	default:
		return "[LOG]"
	case TYPE_PANIC:
		return "[PANIC]"
	case TYPE_ERROR:
		return "[ERROR]"
	case TYPE_WARN:
		return "[WARN]"
	case TYPE_INFO:
		return "[INFO]"
	case TYPE_DEBUG:
		return "[DEBUG]"
	}
}

// Set 设置 log 级别
func (l *LogLevel) Set(v LogLevel) {
	atomic.StoreInt64((*int64)(l), int64(v))
}

// LogTypeAllow 是否允许这个级别的log打印处理
// 比如 LEVEL_INFO 级别是 0111, 那么 TYPE_INFO(0100),TYPE_WARN(0010)
// 都能打印出来.
func (l *LogLevel) LogTypeAllow(m LogType) bool {
	v := atomic.LoadInt64((*int64)(l))
	return (v & int64(m)) != 0
}

type nopCloser struct {
	io.Writer
}

func (*nopCloser) Close() error {
	return nil
}

// NopCloser 创建一个假关闭writer
// nopCloser 假关闭, 一般用于 os.stdin, os.stderr
func NopCloser(w io.Writer) io.WriteCloser {
	return &nopCloser{w}
}

// Logger 在原生 log 之上的封装
type Logger struct {
	// mu 包含数据
	mu sync.Mutex
	// out 输出
	out io.WriteCloser
	// log 原始 log
	log *log.Logger
	// level 日志的级别
	level LogLevel
	// trace 跟踪的级别
	trace LogLevel
	// 用途: 比如将传入的数据发送到阿里云的日志服务
	callback func(*Placehoder, string)
}

func (l *Logger) output(transaction *Placehoder, traceSkip int, err error, t LogType, logMsg string) error {
	var stack Stack
	// 如果开启跟踪
	if l.isTraceEnabled(t) {
		stack = TraceN(traceSkip+1, 32)
	}

	var b bytes.Buffer
	_, _ = fmt.Fprint(&b, t, " ", logMsg)

	if len(logMsg) == 0 || logMsg[len(logMsg)-1] != '\n' {
		_, _ = fmt.Fprint(&b, "\n")
	}

	if err != nil {
		_, _ = fmt.Fprint(&b, "[error]: ", err.Error(), "\n")
		if stack := GetStack(err); stack != nil {
			_, _ = fmt.Fprint(&b, stack.StringWithIndent(1))
		}
	}

	if len(stack) != 0 {
		_, _ = fmt.Fprint(&b, "[stack]: \n", stack.StringWithIndent(1))
	}

	logMsg = b.String()

	l.mu.Lock()
	if l.callback != nil {
		r := Caller(2)
		l.callback(transaction, fmt.Sprintf("%s:%d: %s", filepath.Base(r.File), r.Line, logMsg))
	}
	err = l.log.Output(traceSkip+2, logMsg)
	l.mu.Unlock()
	return err
}

func (l *Logger) SetCallback(callback func(*Placehoder, string)) {
	l.mu.Lock()
	l.callback = callback
	l.mu.Unlock()
}

func New(writer io.Writer, prefix string) *Logger {
	out, ok := writer.(io.WriteCloser)
	if !ok {
		out = NopCloser(writer)
	}
	return &Logger{
		out:   out,
		log:   log.New(out, prefix, LstdFlags|Lshortfile),
		level: LEVEL_DEBUG,
		trace: LEVEL_ERROR,
	}
}

func (l *Logger) Flags() int {
	return l.log.Flags()
}

func (l *Logger) Prefix() string {
	return l.log.Prefix()
}

func (l *Logger) SetFlags(flags int) {
	l.log.SetFlags(flags)
}

func (l *Logger) SetPrefix(prefix string) {
	l.log.SetPrefix(prefix)
}

func (l *Logger) SetLogLevel(v LogLevel) {
	l.level.Set(v)
}

func (l *Logger) SetTraceLevel(v LogLevel) {
	l.trace.Set(v)
}

func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.out.Close()
}

// isLogDisabled 是否关闭日志, TYPE_PANIC 无法被关闭
func (l *Logger) isLogDisabled(t LogType) bool {
	return t != TYPE_PANIC && !l.level.LogTypeAllow(t)
}

// isTraceEnabled 是否跟踪, TYPE_PANIC 一直被跟踪
func (l *Logger) isTraceEnabled(t LogType) bool {
	return t == TYPE_PANIC || l.trace.LogTypeAllow(t)
}

func (l *Logger) Panic(v ...interface{}) {
	t := TYPE_PANIC
	s := fmt.Sprint(v...)
	_ = l.output(nil, 1, nil, t, s)
	os.Exit(1)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	t := TYPE_PANIC
	s := fmt.Sprintf(format, v...)
	_ = l.output(nil, 1, nil, t, s)
	os.Exit(1)
}

func (l *Logger) PanicError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_PANIC
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
		os.Exit(1)
	}
}

func (l *Logger) PanicErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_PANIC
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
		os.Exit(1)
	}
}

func (l *Logger) Error(v ...interface{}) {
	t := TYPE_ERROR
	if l.isLogDisabled(t) {
		return
	}
	s := fmt.Sprint(v...)
	_ = l.output(nil, 1, nil, t, s)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	t := TYPE_ERROR
	if l.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = l.output(nil, 1, nil, t, s)
}

func (l *Logger) ErrorError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_ERROR
		if l.isLogDisabled(t) {
			return
		}
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
	}
}

func (l *Logger) ErrorErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_ERROR
		if l.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
	}
}

func (l *Logger) Warn(v ...interface{}) {
	t := TYPE_WARN
	if l.isLogDisabled(t) {
		return
	}
	s := fmt.Sprint(v...)
	_ = l.output(nil, 1, nil, t, s)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	t := TYPE_WARN
	if l.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = l.output(nil, 1, nil, t, s)
}

func (l *Logger) WarnError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_WARN
		if l.isLogDisabled(t) {
			return
		}
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
	}
}

func (l *Logger) WarnErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_WARN
		if l.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
	}
}

func (l *Logger) Info(v ...interface{}) {
	t := TYPE_INFO
	if l.isLogDisabled(t) {
		return
	}
	s := fmt.Sprint(v...)
	_ = l.output(nil, 1, nil, t, s)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	t := TYPE_INFO
	if l.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = l.output(nil, 1, nil, t, s)
}

func (l *Logger) InfoError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_INFO
		if l.isLogDisabled(t) {
			return
		}
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
	}
}

func (l *Logger) InfoErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_INFO
		if l.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
	}
}

func (l *Logger) Debug(v ...interface{}) {
	t := TYPE_DEBUG
	if l.isLogDisabled(t) {
		return
	}
	s := fmt.Sprint(v...)
	_ = l.output(nil, 1, nil, t, s)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	t := TYPE_DEBUG
	if l.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = l.output(nil, 1, nil, t, s)
}

func (l *Logger) DebugError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_DEBUG
		if l.isLogDisabled(t) {
			return
		}
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
	}
}

func (l *Logger) DebugErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_DEBUG
		if l.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = l.output(nil, 1, err, t, s)
	}
}

func (logger *Logger) SetLogLevelStr(level string) {
	level = strings.ToLower(level)
	var l = LEVEL_INFO
	switch level {
	case "error":
		l = LEVEL_ERROR
	case "warn", "warning":
		l = LEVEL_WARN
	case "debug":
		l = LEVEL_DEBUG
	case "info":
		fallthrough
	default:
		level = "info"
		l = LEVEL_INFO
	}
	logger.SetLogLevel(l)
	Infof("set log level to %s", level)
}

// ------------------------------------------log 包使用----------------------------------------------------

func SetLogLevelStr(level string) {
	level = strings.ToLower(level)
	var l = LEVEL_INFO
	switch level {
	case "error":
		l = LEVEL_ERROR
	case "warn", "warning":
		l = LEVEL_WARN
	case "debug":
		l = LEVEL_DEBUG
	case "info":
		fallthrough
	default:
		level = "info"
		l = LEVEL_INFO
	}
	SetLogLevel(l)
	Infof("set log level to %s", level)
}

// StdLog 默认指向 os.Stderr
// stderr 单元测试中能看到.
var StdLog = New(NopCloser(os.Stderr), "")

func Flags() int {
	return StdLog.log.Flags()
}

func Prefix() string {
	return StdLog.log.Prefix()
}

func SetFlags(flags int) {
	StdLog.log.SetFlags(flags)
}

func SetPrefix(prefix string) {
	StdLog.log.SetPrefix(prefix)
}

func SetLogLevel(v LogLevel) {
	StdLog.level.Set(v)
}

func SetLogTrace(v LogLevel) {
	StdLog.trace.Set(v)
}

func Panic(v ...interface{}) {
	t := TYPE_PANIC
	s := fmt.Sprint(v...)
	_ = StdLog.output(nil, 1, nil, t, s)
	os.Exit(1)
}

func Panicf(format string, v ...interface{}) {
	t := TYPE_PANIC
	s := fmt.Sprintf(format, v...)
	_ = StdLog.output(nil, 1, nil, t, s)
	os.Exit(1)
}

func PanicError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_PANIC
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
		os.Exit(1)
	}
}

func PanicErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_PANIC
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
		os.Exit(1)
	}
}

func Error(v ...interface{}) {
	t := TYPE_ERROR
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprint(v...)
	_ = StdLog.output(nil, 1, nil, t, s)
}

func Errorf(format string, v ...interface{}) {
	t := TYPE_ERROR
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = StdLog.output(nil, 1, nil, t, s)
}

func ErrorError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_ERROR
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
	}
}

func ErrorErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_ERROR
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
	}
}

func TxErrorErrorf(transaction Placehoder, err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_ERROR
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		_ = StdLog.output(&transaction, 1, err, t, s)
	}
}

func Warn(v ...interface{}) {
	t := TYPE_WARN
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprint(v...)
	_ = StdLog.output(nil, 1, nil, t, s)
}

func Warnf(format string, v ...interface{}) {
	t := TYPE_WARN
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = StdLog.output(nil, 1, nil, t, s)
}

func TxWarnf(transaction Placehoder, format string, v ...interface{}) {
	t := TYPE_WARN
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = StdLog.output(&transaction, 1, nil, t, s)
}

func WarnError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_WARN
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
	}
}

func WarnErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_WARN
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
	}
}

func Info(v ...interface{}) {
	t := TYPE_INFO
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprint(v...)
	_ = StdLog.output(nil, 1, nil, t, s)
}

func Infof(format string, v ...interface{}) {
	t := TYPE_INFO
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = StdLog.output(nil, 1, nil, t, s)
}

func TxInfof(transaction Placehoder, format string, v ...interface{}) {
	t := TYPE_INFO
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = StdLog.output(&transaction, 1, nil, t, s)
}

func InfoError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_INFO
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
	}
}

func InfoErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_INFO
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
	}
}

func Debug(v ...interface{}) {
	t := TYPE_DEBUG
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprint(v...)
	_ = StdLog.output(nil, 1, nil, t, s)
}

func Debugf(format string, v ...interface{}) {
	t := TYPE_DEBUG
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = StdLog.output(nil, 1, nil, t, s)
}

func TxDebugf(tx Placehoder, format string, v ...interface{}) {
	t := TYPE_DEBUG
	if StdLog.isLogDisabled(t) {
		return
	}
	s := fmt.Sprintf(format, v...)
	_ = StdLog.output(&tx, 1, nil, t, s)
}

func DebugError(err error, v ...interface{}) {
	if err != nil {
		t := TYPE_DEBUG
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprint(v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
	}
}

func DebugErrorf(err error, format string, v ...interface{}) {
	if err != nil {
		t := TYPE_DEBUG
		if StdLog.isLogDisabled(t) {
			return
		}
		s := fmt.Sprintf(format, v...)
		if myerr, ok := err.(*TracedError); ok {
			s += myerr.Extra
		}
		_ = StdLog.output(nil, 1, err, t, s)
	}
}
