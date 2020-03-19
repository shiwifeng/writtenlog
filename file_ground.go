package writtenlog

// 日志写入文件(Log write file)
import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type FileLogger struct {
	Level       LogLevel     // 日志级别
	filePath    string       // 日志文件路径
	fileName    string       // 日志文件名
	fileObj     *os.File     // 日志对象
	errFileObj  *os.File     // error级别以上日志对象
	maxFileSize int64        //文件大小，字节
	logChon     chan *logMsg //异步写入通道
}

var (
	MaxSize int          = 50000 // 日志通道缓冲区大小
	LogSize int                  // 日志数量
	rwlock  sync.RWMutex         // 并发锁
)

// 异步写入日志结构体
type logMsg struct {
	level     LogLevel //日志级别
	msg       string   //日志内容
	fileName  string   // 日志文件名
	funcName  string   //函数名
	timestamp string   // 时间戳
	line      int      // 异常行号
}

// FileLogger构造函数
// levelStr 错误级别字符串
// fp 文件路径
// fn 文件名称
// maxSize 文件大小，单位：字节
func NewFileLogger(levelStr, fp, fn string, maxSize int64) *FileLogger {
	level, err := parseLogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	f1 := &FileLogger{
		Level:       level,
		filePath:    fp,
		fileName:    fn,
		maxFileSize: maxSize,
		logChon:     make(chan *logMsg, MaxSize),
	}
	err = f1.initFile()
	if err != nil {
		panic(err)
	}
	return f1
}

// 初始化日志文件
func (f *FileLogger) initFile() error {
	// 拼接文件路径
	fullFileName := path.Join(f.filePath, f.fileName)
	// 创建日志文件
	fileObj, err := os.OpenFile(fullFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("initFile os.OpenFile() log failed,err:%v\n", err)
		return err
	}
	// 创建error级别日志文件
	sci := strings.Split(fullFileName, ".")
	sci1 := make([]string, len(sci))
	copy(sci1, sci)
	if len(sci) > 1 {
		sci1 = sci1[:len(sci)-1]
		sci1 = append(sci1, "_err.")
		sci1 = append(sci1, sci[len(sci)-1])
	} else {
		sci1 = append(sci1, "_err")
	}
	fullErrorFileName := strings.Join(sci1, "")
	errFileObj, err := os.OpenFile(fullErrorFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("initFile os.OpenFile() .err file failed,err:%v\n", err)
		return err
	}
	// 打开的日志对象赋值FileLogger结构体
	f.fileObj = fileObj
	f.errFileObj = errFileObj

	// 消费通道，异步写入日志
	for i := 1; i <= 5; i++ {
		go f.consumerLogChan()
	}
	return nil
}

// 判断日志级别
func (f *FileLogger) enable(logLevel LogLevel) bool {
	return logLevel >= f.Level
}

// 判断日志文件是否大于等于阈值
func (f *FileLogger) checkSize(file *os.File) bool {
	// 获取文件对象信息
	fileInfo, err := file.Stat()
	if err != nil {
		//fmt.Printf("checkSize get file size info failed,err:%v\n", err)
		return false
	}
	return fileInfo.Size() >= f.maxFileSize
}

// 切割日志
func (f *FileLogger) splitFile(file *os.File) (*os.File, error) {
	// 需要切割的日志文件
	// 1. 拿到当前的日志文件完整路径
	fileInfo, err := file.Stat()
	if err != nil {
		//fmt.Printf("splitFile get file info failed,err:%v\n", err)
		return nil, err
	}
	// 2. 关闭当前的日志文件
	file.Close()
	logName := path.Join(f.filePath, fileInfo.Name())
	// 3. 备份 xxx.log -> xxx.log.bak20191023133040.0000000
	nowStr := fmt.Sprintf("%s%d", time.Now().Format("20060102150405.00000000"), RandInt64(11, 99))
	// 4. 拼接备份日志文件
	newlogName := fmt.Sprintf("%s.bak%s", logName, nowStr)
	err = os.Rename(logName, newlogName)
	if err != nil {
		fmt.Printf("Rename file failed,err:%v\n", err)
		return nil, err
	}
	// 5. 打开新的日志文件
	fileObj, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Open new file log failed,err:%v\n", err)
		return nil, err
	}
	// 6. 返回新打开的日志文件对象
	return fileObj, nil
}

// 写入文件
func (f *FileLogger) writeLogBackGround(logTmp *logMsg) error {
	if f.checkSize(f.fileObj) {
		nweFileObj, err := f.splitFile(f.fileObj)
		if err != nil {
			select {
			case f.logChon <- logTmp:
			default:
				f.logChon <- logTmp
			}
			return err
		}
		f.fileObj = nweFileObj
	}
	msgInfo := fmt.Sprintf("[%s] [%s] [%s:%s:%d] %s\n", logTmp.timestamp, getLogString(logTmp.level), logTmp.fileName, logTmp.funcName, logTmp.line, logTmp.msg)
	// 写入日志
	_, err := fmt.Fprintf(f.fileObj, msgInfo)
	if err != nil {
		select {
		case f.logChon <- logTmp:
		default:
			f.logChon <- logTmp
		}
		return err
	}
	// 写入error级别以上日志
	if logTmp.level >= ERROR {
		if f.checkSize(f.errFileObj) {
			nweFileObj, err := f.splitFile(f.errFileObj)
			if err != nil {
				return err
			}
			f.errFileObj = nweFileObj
		}
		fmt.Fprintf(f.errFileObj, msgInfo)
	}
	rwlock.Lock()
	LogSize--
	rwlock.Unlock()
	return nil
}

// 处理通道消息
func (f *FileLogger) consumerLogChan() {
	var logTmp *logMsg
	for {
		select {
		case logTmp = <-f.logChon:
			err := f.writeLogBackGround(logTmp)
			if err != nil {
				continue
			}
		default:
			// 取不到日志内容，休息500毫秒
			time.Sleep(time.Millisecond * 300)
		}
	}
}

// 格式化日志信息
func (f *FileLogger) log(lv LogLevel, format string, a ...interface{}) {
	if f.enable(lv) {
		msg := fmt.Sprintf(format, a...)
		now := time.Now()
		funcName, fileName, lineNo := getInfo(3)
		//1. 先把日志发送到通道
		//1.1 创建 logMsg 对象
		logTamp := &logMsg{
			level:     lv,
			msg:       msg,
			funcName:  funcName,
			fileName:  fileName,
			timestamp: now.Format("2006-01-02 15:04:05"),
			line:      lineNo,
		}
		select {
		case f.logChon <- logTamp:
			rwlock.Lock()
			LogSize++
			rwlock.Unlock()
		default:
			fmt.Printf("send info to logChon failed\n")
		}
	}
}

// 关闭日志文件对象
func (f *FileLogger) Close() {
	f.fileObj.Close()
	f.errFileObj.Close()
}

// Debug 级别
func (f *FileLogger) Debug(format string, a ...interface{}) {
	f.log(DEBUG, format, a...)
}

// Trace 级别
func (f *FileLogger) Trace(format string, a ...interface{}) {
	f.log(TRACE, format, a...)
}

// Info 级别
func (f *FileLogger) Info(format string, a ...interface{}) {
	f.log(INFO, format, a...)
}

// Warning 级别
func (f *FileLogger) Warning(format string, a ...interface{}) {
	f.log(WARNING, format, a...)
}

// Error 级别
func (f *FileLogger) Error(format string, a ...interface{}) {
	f.log(ERROR, format, a...)
}

// Fatal 级别
func (f *FileLogger) Fatal(format string, a ...interface{}) {
	f.log(FATAL, format, a...)
}
