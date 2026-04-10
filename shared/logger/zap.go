package logger

import (
	"fmt"
	"live-interact-engine/shared/env"
	"log"
	"os"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapConfig struct {
	level      string
	fileName   string
	maxSize    int
	maxAge     int
	maxBackups int
}

var (
	config     zapConfig
	initOnce   sync.Once
	configLock sync.RWMutex
)

func UpdateConfig() {
	config = zapConfig{
		level:      env.GetString("LOG_LEVEL", "info"),
		fileName:   env.GetString("LOG_FILENAME", "./logs/app.log"),
		maxSize:    env.GetInt("LOG_MAX_SIZE", 100),
		maxAge:     env.GetInt("LOG_MAX_AGE", 7),
		maxBackups: env.GetInt("LOG_MAX_BACKUPS", 10),
	}

	if config.maxSize <= 0 {
		config.maxSize = 100 // 默认 100MB
	}
	if config.maxAge <= 0 {
		config.maxAge = 7 // 默认保留 7 天
	}
	if config.maxBackups <= 0 {
		config.maxBackups = 10 // 默认保留 10 个备份
	}
}

func init() {
	var result error

	initOnce.Do(func() {
		// 步骤 1: 更新配置
		UpdateConfig()

		// 步骤 2: 获取日志写入器
		writeSyncer := getLogWriter()
		encoder := getEncoder()

		// 步骤 3: 解析日志级别
		var l = new(zapcore.Level)
		err := l.UnmarshalText([]byte(config.level))
		if err != nil {
			result = fmt.Errorf("invalid log level '%s': %w", config.level, err)
			return // 注意这里是 return，不是 return result
		}

		// 步骤 4: 创建核心 logger
		core := zapcore.NewCore(encoder, writeSyncer, l)
		lg := zap.New(core, zap.AddCaller())

		// 步骤 5: 全局替换
		zap.ReplaceGlobals(lg)

		log.Println("初始化zap成功")
	})

	if result != nil {
		panic(result)
	}

	log.Println("logger of zap init success")
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   config.fileName,
		MaxSize:    config.maxSize,
		MaxBackups: config.maxBackups,
		MaxAge:     config.maxAge,
	}

	// 添加文件写入器
	writers := []zapcore.WriteSyncer{zapcore.AddSync(lumberJackLogger)}

	writers = append(writers, zapcore.AddSync(os.Stdout))

	return zapcore.NewMultiWriteSyncer(writers...)
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 - 15:04:05"))
}
