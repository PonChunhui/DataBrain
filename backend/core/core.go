package core

import (
	"fmt"
	"os"

	"devops-backend/global"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitConfig() {
	v := viper.New()
	v.SetConfigFile("config.yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	if err := v.Unmarshal(&global.GVA_CONFIG); err != nil {
		panic(fmt.Errorf("Fatal error unmarshal config: %s \n", err))
	}
}

func InitLog() {
	var level zapcore.Level
	switch global.GVA_CONFIG.Log.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  global.GVA_CONFIG.Log.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	var encoder zapcore.Encoder
	if global.GVA_CONFIG.Log.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)

	logger := zap.New(core, zap.AddCaller())
	if global.GVA_CONFIG.Log.ShowLine {
		logger = logger.WithOptions(zap.AddCallerSkip(1))
	}

	global.GVA_LOG = logger
}
