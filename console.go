package writtenlog

// 终端输出日志(Terminal output log)
import (
	"fmt"
	"time"
)

// 定义日志级别结构体
type ConsoleLogger struct {
	Level LogLevel
}

// 初始化日志级别结构体
func NewConsoleLog(levelStr string) ConsoleLogger {
	level, err := parseLogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	return ConsoleLogger{Level: level}
}

// 判断日志级别
func (c ConsoleLogger) enable(logLevel LogLevel) bool {
	return logLevel >= c.Level
}

// 格式化日志信息
func (c ConsoleLogger) log(lv LogLevel, format string, a ...interface{}) {
	if c.enable(lv) {
		msg := fmt.Sprintf(format, a...)
		now := time.Now()
		funcName, fileName, lineNo := getInfo(3)
		fmt.Printf("[%s] [%s] [%s:%s:%d] %s\n", now.Format("2006-01-02 15:04:05"), getLogString(lv), fileName, funcName, lineNo, msg)
	}
}

// Debug 级别
func (c ConsoleLogger) Debug(format string, a ...interface{}) {
	c.log(DEBUG, format, a...)
}

// Trace 级别
func (c ConsoleLogger) Trace(format string, a ...interface{}) {
	c.log(TRACE, format, a...)
}

// Info 级别
func (c ConsoleLogger) Info(format string, a ...interface{}) {
	c.log(INFO, format, a...)
}

// Warning 级别
func (c ConsoleLogger) Warning(format string, a ...interface{}) {
	c.log(WARNING, format, a...)
}

// Error 级别
func (c ConsoleLogger) Error(format string, a ...interface{}) {
	c.log(ERROR, format, a...)
}

// Fatal 级别
func (c ConsoleLogger) Fatal(format string, a ...interface{}) {
	c.log(FATAL, format, a...)
}
