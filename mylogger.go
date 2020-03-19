package writtenlog

// 自定义日志库
import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"path"
	"runtime"
	"strings"
)

// 定义日志级别数据类型
type LogLevel uint16

// 定义日志级别常量
const (
	UNKNOWN LogLevel = iota
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	FATAL
)

// 判断接受的日志级别
func parseLogLevel(s string) (LogLevel, error) {
	s = strings.ToLower(s)
	switch s {
	case "debug":
		return DEBUG, nil
	case "trace":
		return TRACE, nil
	case "info":
		return INFO, nil
	case "warning":
		return WARNING, nil
	case "error":
		return ERROR, nil
	case "fatal":
		return FATAL, nil
	}
	err := errors.New("Invalid log level!\n")
	return UNKNOWN, err
}

// 判断错误级别字符串
func getLogString(lv LogLevel) string {
	switch lv {
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return "UNKNOWN"
}

func getInfo(skip int) (funcName, fileName string, lineNo int) {
	pc, file, lineNo, ok := runtime.Caller(skip) // 获取调用层数信息，返回：文件路径，文件名，行数
	if !ok {
		fmt.Printf("getInfo a runtime.Caller() failed,err:%v\n", ok)
		return
	}
	funcName = runtime.FuncForPC(pc).Name()
	funcName = strings.Split(funcName, ".")[1]
	fileName = path.Base(file)
	return
}

// 范围随机数,包含 min ,不包含 max
// [1-3)
// min 最小值
// max 最大值
func RandInt64(min, max int64) int64 {
	maxBigInt := big.NewInt(max)
	i, _ := rand.Int(rand.Reader, maxBigInt)
	if i.Int64() < min {
		return RandInt64(min, max)
	}
	return i.Int64()
}
