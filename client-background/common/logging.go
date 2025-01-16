package common

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitGlobalLogger() error {
	globalLogfileName := time.Now().UTC().Format(TIME_FORMAT) + ".log"
	globalLog, err := os.OpenFile(filepath.Join(GetClientLogsDir(), globalLogfileName), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return err
	}
	GlobalLogger = CreateLogger(globalLog)

	return nil
}

func CreateLogger(logFile *os.File) *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewConsoleEncoder(config)
	writer := zapcore.Lock(zapcore.AddSync(logFile))

	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewCore(fileEncoder, writer, defaultLogLevel)

	logger := zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel))

	return logger
}

func InitLogger(filePath string) (*zap.Logger, error) {
	logFile, err := os.Create(filePath)

	if err != nil {
		return nil, err
	}

	logger := CreateLogger(logFile)

	return logger, nil
}
