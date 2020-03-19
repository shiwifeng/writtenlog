## 异步写日志库
可设置日志切割大小，默认：5M（Log cutting size can be set, default:5M）

### 接口
- NewFileLogger(levelStr, fp, fn string, maxSize int64) *FileLogger // 初始化
- Debug(format string, a ...interface{})    // Debug 日志级别(The level of logging)
- Trace(format string, a ...interface{})    // Trace 日志级别(The level of logging)
- Info(format string, a ...interface{}) // Info 日志级别(The level of logging)
- Warning(format string, a ...interface{}) // Warning 日志级别(The level of logging)
- Error(format string, a ...interface{}) // Error 日志级别(The level of logging)
- Fatal(format string, a ...interface{}) // Fatal 日志级别(The level of logging)

### 示例
```
goos: windows
goarch: amd64
=== RUN   TestFileLogger_Debug
[用时:1.133352 Second] id:50000
[用时:1.312250 Second] id:100000
[用时:1.420189 Second] id:150000
[用时:1.382209 Second] id:200000
[用时:1.296259 Second] id:250000
--- PASS: TestFileLogger_Debug (12.55s)
PASS
```
```
package writtenlog

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestFileLogger_Debug(t *testing.T) {
	i := 1
	num := 50000 // 如果是异步写写需要配合缓冲大小
	n := 5       //并发次数
	var now1 time.Time
	var now2 time.Time
	now1 = time.Now()
	log := NewFileLogger("debug", ".", "my.log", 5<<20)
	str := "###################################################################################################################################################################################################################"
	tmpLogSize := 0
	for {
		if i > (num * n) {
			// 日志全部写入文件，跳出循环
			if tmpLogSize == LogSize {
				break
			}
			time.Sleep(time.Second)
			tmpLogSize = LogSize
			continue
		}
		rand.Seed(time.Now().UnixNano() + int64(i)) //随机种子
		randNum := rand.Intn(7)
		if randNum < 1 {
			randNum++
		}
		switch randNum {
		case int(DEBUG):
			log.Debug("id:%d This is a debug log...%s", i, str)
		case int(TRACE):
			log.Trace("id:%d This is a trace log...%s", i, str)
		case int(INFO):
			log.Info("id:%d This is a info log...%s", i, str)
		case int(WARNING):
			log.Warning("id:%d This is a warning log...%s", i, str)
		case int(ERROR):
			log.Error("id:%d This is a error log...%s", i, str)
		case int(FATAL):
			log.Fatal("id:%d This is a fatal log...%s", i, str)
		}
		if i%num*n == 0 {
			now2 = time.Now()
			fmt.Printf("[用时:%f Second] id:%d\n", now2.Sub(now1).Seconds(), i)
			time.Sleep(time.Second)
			if num*n != i {
				now1 = time.Now()
			}
		}
		i++
	}
	log.Close()
}

```