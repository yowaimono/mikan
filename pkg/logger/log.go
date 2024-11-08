package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	logger     *log.Logger
	level      string
	outputType int // 使用枚举类型
	fileLogger *FileLogger
}

var (
	instanceLog *Logger
	olog        sync.Once
)

func GetLogger() *Logger {
	olog.Do(func() {
		instanceLog = &Logger{
			logger:     log.New(os.Stdout, "", 0),
			level:      "INFO", // 默认日志级别为 INFO
			outputType: CONSOLE,
		}
	})
	return instanceLog
}

func (l *Logger) SetOutputType(outputType int, filename ...string) error {
	l.outputType = outputType
	if outputType&FILE != 0 {
		if len(filename) == 0 {
			return fmt.Errorf("filename is required for file output")
		}
		fileLogger, err := NewFileLogger(filename[0], 1024*1024)
		if err != nil {
			return err
		}
		l.fileLogger = fileLogger
	}
	return nil
}

func (l *Logger) Close() {
	if l.fileLogger != nil {
		l.fileLogger.Close()
	}
}

// func contains(slice []string, value string) bool {
// 	for _, v := range slice {
// 		if v == value {
// 			return true
// 		}
// 	}
// 	return false
// }

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level string) {
	l.level = level
}

// logf 是日志记录的通用函数
func (l *Logger) logf(level, format string, v ...interface{}) {
	if l.shouldLog(level) {
		now := time.Now().Format("2006-01-02 15:04:05")

		_, file, line, _ := runtime.Caller(2)
		pkname := path.Base(path.Dir(file))
		file = file[strings.LastIndex(file, "/")+1:]
		message := fmt.Sprintf(format, v...)

		pc, _, _, _ := runtime.Caller(2)
		function := runtime.FuncForPC(pc).Name()
		function = function[strings.LastIndex(function, ".")+1:]

		var levelColor string
		switch level {
		case "INFO":
			levelColor = "\033[32m" // 绿色
		case "ERROR":
			levelColor = "\033[31m" // 红色
		case "WARN":
			levelColor = "\033[33m" // 黄色
		case "DEBUG":
			levelColor = "\033[34m" // 青色
		case "FATAL":
			levelColor = "\033[35m" // 紫色
		default:
			levelColor = "\033[0m" // 默认颜色
		}

		logMessage := fmt.Sprintf("[\033[35mMIN\033[0m] [\033[34m%s\033[0m] [\033[36m%s/%s:%d -> %s\033[0m] [\033[33m%s%s\033[0m] %s", now, pkname, file, line, function, levelColor, level, message)

		fileMessage := fmt.Sprintf("[%s] [%s/%s:%d -> %s] [%s] %s\n", now, pkname, file, line, function, level, message)

		// 检查 format 是否包含格式化动词
		// if strings.Contains(format, "%") {
		// 	logMessage += fmt.Sprintf(format, v...)
		// } else {
		// 	logMessage += format
		// }

		if l.outputType == CONSOLE || l.outputType == CONSOLE_AND_FILE {
			l.logger.Println(logMessage)
		}
		if l.outputType == FILE || l.outputType == CONSOLE_AND_FILE {
			l.fileLogger.Write(fileMessage)
		}
	}
}

// shouldLog 判断是否应该记录日志
func (l *Logger) shouldLog(level string) bool {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	levelIndex := indexOf(levels, level)
	currentLevelIndex := indexOf(levels, l.level)
	return levelIndex >= currentLevelIndex
}

// indexOf 返回元素在切片中的索引，如果不存在则返回 -1
func indexOf(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// Info 记录信息级别的日志
func Info(format string, v ...interface{}) {
	GetLogger().logf("INFO", format, v...)
}

// Error 记录错误级别的日志
func Error(format string, v ...interface{}) {
	GetLogger().logf("ERROR", format, v...)
}

// Warn 记录警告级别的日志
func Warn(format string, v ...interface{}) {
	GetLogger().logf("WARN", format, v...)
}

// Debug 记录调试级别的日志
func Debug(format string, v ...interface{}) {
	GetLogger().logf("DEBUG", format, v...)
}

// Fatal 记录致命错误级别的日志并退出程序
func Fatal(format string, v ...interface{}) {
	GetLogger().logf("FATAL", format, v...)
	os.Exit(1)
}

const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

const (
	CONSOLE = 1 << iota
	FILE
)

const (
	CONSOLE_AND_FILE = CONSOLE | FILE
)
