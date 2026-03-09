package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L     *zap.Logger
	Sugar *zap.SugaredLogger
)

// InitLogger 初始化日志器
func InitLogger(env string) error {
	var (
		log *zap.Logger
		err error
	)

	switch env {
	case "production", "prod":
		// 生产环境：JSON格式，性能优先
		log, err = zap.NewProduction(
			zap.AddCaller(),                   // 添加调用者信息
			zap.AddCallerSkip(1),              // 跳过一层调用
			zap.AddStacktrace(zap.ErrorLevel), // 错误级别及以上记录堆栈
			zap.Fields( // 添加全局字段
				zap.String("env", env),
				zap.String("service", "your-service-name"),
			),
		)

	case "staging", "test":
		// 预发/测试环境：JSON格式但带更多信息
		log, err = zap.NewProduction(
			zap.AddCaller(),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zap.WarnLevel), // 警告级别就记录堆栈
			zap.Development(),                // 开启开发模式特性
			zap.Fields(
				zap.String("env", env),
			),
		)

	default:
		// 开发环境：控制台彩色输出，易读
		config := zap.NewDevelopmentConfig()

		// 开发环境编码器配置
		config.EncoderConfig = zapcore.EncoderConfig{
			// 关键字段 - 控制显示的内容
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey, // 生产环境可省略以减小体积
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,

			// 编码方式
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // 更友好的时间格式
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,       // 短路径格式
			EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 内置彩色级别
			EncodeName:     zapcore.FullNameEncoder,
		}

		// 开发环境设置
		config.Development = true
		config.DisableCaller = false     // 显示调用者
		config.DisableStacktrace = false // 不禁止堆栈
		config.Sampling = nil            // 开发环境不采样
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel) // 开发环境默认调试级别

		// 应用配置
		log, err = config.Build(
			zap.AddCallerSkip(1),
			zap.WithCaller(true),
		)
	}

	if err != nil {
		return err
	}

	// 设置全局 logger
	L = log
	Sugar = L.Sugar()

	// 使用更结构化的日志记录初始化信息
	L.Info("Logger initialized successfully",
		zap.String("environment", env),
		zap.String("log_format", getLogFormat(env)),
		zap.Bool("has_caller", true),
	)

	return nil
}

// getLogFormat 获取日志格式描述
func getLogFormat(env string) string {
	switch env {
	case "production", "prod", "staging", "test":
		return "json"
	default:
		return "console"
	}
}

// WithContext 添加上下文信息（常用于请求链路的追踪）
func WithContext(ctx context.Context) *zap.Logger {
	// 这里可以添加从context中提取的追踪信息
	// 例如：traceID、spanID、userID等

	// 示例：假设context中有requestID
	/*
		if requestID, ok := ctx.Value("requestID").(string); ok {
			return L.With(zap.String("request_id", requestID))
		}
	*/
	return L
}

// Sync 同步日志
func Sync() error {
	if L != nil {
		return L.Sync()
	}
	return nil
}

// GetLogger 获取原始logger
func GetLogger() *zap.Logger {
	return L
}

// GetSugar 获取sugar logger
func GetSugar() *zap.SugaredLogger {
	return Sugar
}
