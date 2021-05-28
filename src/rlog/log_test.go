package rlog

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTimeNoRollingLog(t *testing.T) {
	pPath := filepath.Join(os.TempDir(), "rlog")
	os.MkdirAll(pPath, 0777)
	dirPath, e := ioutil.TempDir(pPath, "rlog-")
	MustNoError(e)

	t.Log(dirPath)
	f, e := NewNoRollingFile(filepath.Join(dirPath, "mylogfortest"))
	MustNoError(e)
	log.SetOutput(f)

	for i := 0; i < 3; i++ {
		log.Printf("i=%d, sdfsadf", i)
		time.Sleep(time.Second)
	}
}

func TestTimeRollingLog(t *testing.T) {
	pPath := filepath.Join(os.TempDir(), "rlog")
	os.MkdirAll(pPath, 0777)
	dirPath, e := ioutil.TempDir(pPath, "rlog-")
	MustNoError(e)
	log.Println(dirPath)

	f, e := NewRollingTimeFile(filepath.Join(dirPath, "mylogfortest.log"), "* * * * *")
	MustNoError(e)
	log.SetOutput(f)

	for i := 0; i < 3000000; i++ {
		log.Printf("i=%d, sdfsadf", i)
		time.Sleep(time.Second)
	}
}

func TestRollingLog(t *testing.T) {
	// 创建临时目录作为测试
	pPath := filepath.Join(os.TempDir(), "rlog")
	os.MkdirAll(pPath, 0777)
	dirPath, e := ioutil.TempDir(pPath, "rlog-")
	MustNoError(e)

	t.Logf("临时目录为: %s", dirPath)

	// 创建一个根据指定大小回滚的文件.
	f, e := NewRollingFile(filepath.Join(dirPath, "mylogfortest"), 30, 10)
	MustNoError(e)

	// 设置golang 自带的log的输出到f
	log.SetOutput(f)

	// 设置rlog的默认输出也到f
	StdLog = New(f, "")

	log.Printf("sdfsadf")
	log.Printf("sdfsadf")
	log.Printf("sdfsadf")
	log.Printf("sdfsadf")
	Info("sdfsadf")
	Info("sdfsadf")
	Info("sdfsadf")
}

func TestLevel(t *testing.T) {
	Help_SetLogOutput(t)
	log.Printf("%0.4b\n", LEVEL_NONE)
	log.Printf("%0.4b\n", LEVEL_ERROR)
	log.Printf("%0.4b\n", LEVEL_WARN)
	log.Printf("%0.4b\n", LEVEL_INFO)
	log.Printf("%0.4b\n", LEVEL_DEBUG)
	log.Println("......")
	log.Printf("%0.4b\n", TYPE_PANIC)
	log.Printf("%0.4b\n", TYPE_ERROR)
	log.Printf("%0.4b\n", TYPE_WARN)
	log.Printf("%0.4b\n", TYPE_INFO)
	log.Printf("%0.4b\n", TYPE_DEBUG)
}

func TestLog(t *testing.T) {
	Help_SetLogOutput(t)
	// 创建一个log级别
	l := LogLevel(LEVEL_INFO)
	ret := l.LogTypeAllow(TYPE_ERROR)
	Must(ret)
	ret = l.LogTypeAllow(TYPE_DEBUG)
	Must(!ret)
}

func TestLogger(t *testing.T) {
	Help_SetLogOutput(t)
	SetLogLevel(LEVEL_DEBUG)
	SetLogTrace(LEVEL_DEBUG)
	SetPrefix("wayne")
	SetFlags(LstdFlags)
	Debugf("jjsdfaf,%s", "12233")
	DebugErrorf(NewErrorf("hhhhhh"), "jjsdfaf,%s", "12233")
}

func TestLogger1(t *testing.T) {
	Help_SetLogOutput(t)
	SetLogLevel(LEVEL_DEBUG)
	SetLogTrace(LEVEL_DEBUG)
	SetPrefix("wayne")
	SetFlags(LstdFlags | Lshortfile | Llongfile | Ltime | Lmicroseconds | Ldate)
	log.Println(Prefix())
	log.Println(Flags())
}

func TestLogger2(t *testing.T) {
	Help_SetLogOutput(t)
	DebugErrorf(nil, "")
	DebugError(nil, "")
	Debug()
	Debugf("")

	InfoErrorf(nil, "")
	InfoError(nil, "")
	Info()
	Infof("")

	WarnErrorf(nil, "")
	WarnError(nil, "")
	Warn()
	Warnf("")

	ErrorErrorf(nil, "")
	ErrorError(nil, "")
	Error()
	Errorf("")

}

func testLogger3(t *testing.T) {
	PanicErrorf(nil, "")
	PanicError(nil, "")
	Panic()
}
