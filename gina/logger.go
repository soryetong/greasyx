package gina

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ILog struct {
	*zap.Logger
}

func initILog() {
	viper.SetDefault("Log.Path", "./logs")
	viper.SetDefault("Log.Mode", "both")
	viper.SetDefault("Log.Recover", false)
	viper.SetDefault("Log.MaxSize", 100)
	viper.SetDefault("Log.MaxBackups", 3)
	viper.SetDefault("Log.MaxAge", 7)
	viper.SetDefault("Log.Compress", true)

	Log = &ILog{
		Logger: zap.New(
			getCore(),
			zap.AddCaller(),
			zap.AddCallerSkip(0),
			zap.AddStacktrace(zap.ErrorLevel),
		),
	}
}

func (l *ILog) With(fields ...zap.Field) *ILog {
	return &ILog{
		Logger: l.Logger.With(fields...),
	}
}

func (l *ILog) WithCtx(ctx context.Context) *ILog {
	var traceIdStr, sourceStr string
	traceId := ctx.Value("trace_id")
	if traceId != nil {
		traceIdStr, _ = traceId.(string)
	}

	source := ctx.Value("source")
	if source != nil {
		sourceStr, _ = source.(string)
	}

	if traceIdStr != "" {
		l.With(zap.String("trace_id", traceIdStr))
	}
	if sourceStr != "" {
		l.With(zap.String("source", sourceStr))
	}

	return l
}

// Core is a minimal, fast logger interface. It's designed for library authors
// to wrap in a more user-friendly API.
// only use infoLevel、errorLevel. want update can change == to > or >= or <= or <
func getCore() zapcore.Core {
	path := viper.GetString("Log.Path")
	mode := viper.GetString("Log.Mode")
	doRecover := viper.GetBool("Log.Recover")

	encoder := zapcore.NewJSONEncoder(getEncoderConfig())
	debugWrite := getLogWriter(path, mode, doRecover, zapcore.DebugLevel)
	infoWrite := getLogWriter(path, mode, doRecover, zapcore.InfoLevel)
	warnWrite := getLogWriter(path, mode, doRecover, zapcore.WarnLevel)
	errorWrite := getLogWriter(path, mode, doRecover, zapcore.ErrorLevel)
	debugLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.DebugLevel
	})
	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	})
	warnLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.WarnLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.ErrorLevel
	})
	return zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(debugWrite), debugLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(infoWrite), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(warnWrite), warnLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWrite), errorLevel),
	)
}

// A WriteSyncer is an io.Writer that can also flush any buffered data. Note
// that *os.File (and thus, os.Stderr and os.Stdout) implement WriteSyncer.
func getLogWriter(path, mode string, recover bool, level zapcore.Level) zapcore.WriteSyncer {
	maxSize := viper.GetInt("Log.MaxSize")
	maxBackups := viper.GetInt("Log.MaxBackups")
	maxAge := viper.GetInt("Log.MaxAge")
	compress := viper.GetBool("Log.Compress")
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	fileName := fmt.Sprintf("%s%s/%s.log", path, time.Now().Format("2006-01-02"), level)
	var fileWriter io.Writer
	if recover {
		fileWriter = NewCustomWrite(fileName, maxSize, maxBackups, maxAge, compress)
	} else {
		fileWriter = &lumberjack.Logger{
			Filename:   fileName,
			MaxSize:    maxSize,    // 单文件最大容量, 单位是MB
			MaxBackups: maxBackups, // 最大保留过期文件个数
			MaxAge:     maxAge,     // 保留过期文件的最大时间间隔, 单位是天
			Compress:   compress,   // 是否需要压缩滚动日志, 使用的gzip压缩
			LocalTime:  true,       // 是否使用计算机的本地时间, 默认UTC
		}
	}
	var writer zapcore.WriteSyncer
	switch mode {
	case "file":
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter))
	case "console":
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	case "close":
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(io.Discard))
	default:
		writer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter))
	}

	return writer
}

// An EncoderConfig allows users to configure the concrete encoders supplied by zap core
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file_line",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

// PrimitiveArrayEncoder is the subset of the ArrayEncoder interface that deals
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// CustomWrite 自定义写入器
type CustomWrite struct {
	mu       sync.Mutex
	filepath string
	logger   *lumberjack.Logger
	inner    zapcore.WriteSyncer
	buffer   *bufio.Writer
	done     chan struct{}
}

// NewCustomWrite 创建自定义写入器
func NewCustomWrite(filepath string, maxSize, maxBackups, maxAge int, compress bool) *CustomWrite {
	cw := &CustomWrite{
		filepath: filepath,
		done:     make(chan struct{}),
	}
	cw.initLogger(filepath, maxSize, maxBackups, maxAge, compress)

	// 启动文件状态监控
	go cw.monitorFile()

	return cw
}

// initLogger 初始化日志文件和写入器
func (cw *CustomWrite) initLogger(filepath string, maxSize, maxBackups, maxAge int, compress bool) {
	cw.logger = &lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}
	cw.inner = zapcore.AddSync(cw.logger)
	cw.buffer = bufio.NewWriterSize(cw.inner, 4096)
}

// Write 写入日志
func (cw *CustomWrite) Write(p []byte) (n int, err error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// 写入缓冲区
	n, err = cw.buffer.Write(p)
	if err != nil {
		cw.recreateLogger()
		return n, err
	}

	// 刷新缓冲区
	err = cw.flushBuffer()
	if err != nil {
		cw.recreateLogger()
	}
	return n, err
}

// recreateLogger 重新创建日志文件和写入器
func (cw *CustomWrite) recreateLogger() {
	cw.logger.Close()
	cw.initLogger(cw.filepath, cw.logger.MaxSize, cw.logger.MaxBackups, cw.logger.MaxAge, cw.logger.Compress)
}

// flushBuffer 刷新缓冲区
func (cw *CustomWrite) flushBuffer() error {
	return cw.buffer.Flush()
}

// monitorFile 异步监控日志文件状态
func (cw *CustomWrite) monitorFile() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cw.checkFile()
		case <-cw.done:
			return
		}
	}
}

// checkFile 检查日志文件是否存在
func (cw *CustomWrite) checkFile() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if _, err := os.Stat(cw.filepath); os.IsNotExist(err) {
		cw.recreateLogger()
	}
}

func (cw *CustomWrite) Sync() error {
	return nil
}

// Close 关闭写入器
func (cw *CustomWrite) Close() error {
	close(cw.done)
	cw.mu.Lock()
	defer cw.mu.Unlock()
	err := cw.flushBuffer()
	cw.logger.Close()
	return err
}
