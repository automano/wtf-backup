package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger 日志记录器结构
type Logger struct {
	level  LogLevel
	output io.Writer
	prefix string
}

// NewLogger 创建新的日志记录器
func NewLogger(level LogLevel, output io.Writer, prefix string) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		level:  level,
		output: output,
		prefix: prefix,
	}
}

// formatMessage 格式化日志消息
func (l *Logger) formatMessage(level string, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [%s] %s: %s\n", timestamp, level, l.prefix, message)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		fmt.Fprint(l.output, l.formatMessage("DEBUG", format, args...))
	}
}

// Info 记录信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		fmt.Fprint(l.output, l.formatMessage("INFO", format, args...))
	}
}

// Warn 记录警告级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		fmt.Fprint(l.output, l.formatMessage("WARN", format, args...))
	}
}

// Error 记录错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		fmt.Fprint(l.output, l.formatMessage("ERROR", format, args...))
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput 设置输出目标
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
}

// SetPrefix 设置日志前缀
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// 全局默认日志记录器
var defaultLogger = NewLogger(LogLevelInfo, os.Stdout, "WTF-Backup")

// SetDefaultLogger 设置默认日志记录器
func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

// Debug 使用默认日志记录器记录调试级别日志
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info 使用默认日志记录器记录信息级别日志
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warn 使用默认日志记录器记录警告级别日志
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Error 使用默认日志记录器记录错误级别日志
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}
