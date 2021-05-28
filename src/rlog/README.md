# rlog
    rlog 纯净无依赖
    rlog 专业打日志的golang模块.

## 使用方法
    创建一个目录: rlog,
    进入 rlog 目录,
    go get github.com/golangtool/rlog

## 功能
### 调用栈跟踪
    跟踪也分级, 跟打日志一样
    debug
    info
    warn
    error
    panic

### 日志分级
    debug
    info
    warn
    error
    panic

## 日志切割
    按照时间切割
    按照大小切割

### 按照时间切割
    下面每分钟切割一次, "* * * * *" 为 crontab 的语法.

    func TestTimeRollingLog(t *testing.T) {
    	pPath := filepath.Join(os.TempDir(), "rlog")
    	os.MkdirAll(pPath, 0777)
    	dirPath, e := ioutil.TempDir(pPath, "rlog-")
    	rlog.MustNoError(e)

    	f, e := rlog.NewRollingTimeFile(filepath.Join(dirPath, "mylogfortest"), "* * * * *")
    	rlog.MustNoError(e)
    	log.SetOutput(f)

    	for i := 0; ; i++ {
    		log.Printf("i=%d, sdfsadf", i)
    		time.Sleep(time.Second)
    	}
    }

    crontab: 分, 时, 号(月份中的号), 月, 星期
    如果需要每 15 分钟切割, crontab字符串写为: */15 * * * *
    如果需要每天零点切割, crontab字符串写为: 0 0 * * *

### 按照大小切割
    下面每1024个字节切割一份, 只保留30份.

    func TestRollingLog(t *testing.T) {
        // 创建临时目录作为测试
        pPath := filepath.Join(os.TempDir(), "rlog")
        os.MkdirAll(pPath, 0777)
        dirPath, e := ioutil.TempDir(pPath, "rlog-")
        rlog.MustNoError(e)

        t.Logf("临时目录为: %s", dirPath)

        // 创建一个根据指定大小回滚的文件.
        f, e := rlog.NewRollingFile(filepath.Join(dirPath, "mylogfortest"), 30, 10)
        rlog.MustNoError(e)

        // 设置golang 自带的log的输出到f
        log.SetOutput(f)

        // 设置rlog的默认输出也到f
        rlog.StdLog = rlog.New(f, "")

        log.Printf("sdfsadf")
        log.Printf("sdfsadf")
        log.Printf("sdfsadf")
        log.Printf("sdfsadf")
        rlog.Info("sdfsadf")
        rlog.Info("sdfsadf")
        rlog.Info("sdfsadf")
    }
