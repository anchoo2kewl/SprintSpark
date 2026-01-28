package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes and returns a configured Zap logger based on environment
func InitLogger(env, logLevel string) (*zap.Logger, error) {
	var config zap.Config

	if env == "production" {
		// Production logger: JSON format, INFO level by default
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Development logger: Console format, DEBUG level by default
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Override log level if specified
	if logLevel != "" {
		level, err := zapcore.ParseLevel(logLevel)
		if err == nil {
			config.Level = zap.NewAtomicLevelAt(level)
		}
	}

	// Build the logger
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// MustInitLogger initializes the logger or exits if it fails
func MustInitLogger(env, logLevel string) *zap.Logger {
	logger, err := InitLogger(env, logLevel)
	if err != nil {
		// Fallback to stderr if logger init fails
		_, _ = os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}
	return logger
}
