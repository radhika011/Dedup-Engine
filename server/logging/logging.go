package logging

import (
	"dedup-server/common"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var GlobalLogger *zap.Logger

func InitGlobalLogger() {
	logFileName := time.Now().UTC().Format(common.TimeFormat) + ".log"
	globalLog, err := os.OpenFile(filepath.Join(common.ServerLogsPath, logFileName), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	GlobalLogger = CreateLogger(globalLog)
}
func CreateLogger(logFile *os.File) *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewConsoleEncoder(config)
	writer := zapcore.Lock(zapcore.AddSync(logFile))

	defaultLogLevel := zapcore.InfoLevel
	core := zapcore.NewCore(fileEncoder, writer, defaultLogLevel)

	logger := zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}
