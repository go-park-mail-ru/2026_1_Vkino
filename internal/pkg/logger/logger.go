package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultLevel  = "info"
	defaultFormat = "text"
)

type Config struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

type Logger struct {
	entry *logrus.Entry
}

type contextKey struct{}

func New(cfg Config) (*Logger, error) {
	cfg = cfg.withDefaults()

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", cfg.Level, err)
	}

	base := logrus.New()
	base.SetLevel(level)

	output, err := buildOutput(cfg.OutputPath)
	if err != nil {
		return nil, err
	}

	base.SetOutput(output)

	switch cfg.Format {
	case "json":
		base.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "text":
		base.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339Nano,
			ForceColors:     cfg.OutputPath == "",
		})
	default:
		return nil, fmt.Errorf("unsupported log format %q", cfg.Format)
	}

	return &Logger{entry: logrus.NewEntry(base)}, nil
}

func ContextWithLogger(ctx context.Context, log *Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if log == nil {
		return ctx
	}

	return context.WithValue(ctx, contextKey{}, log)
}

func FromContext(ctx context.Context) *Logger {
	if ctx != nil {
		log, ok := ctx.Value(contextKey{}).(*Logger)
		if ok && log != nil {
			return log
		}
	}

	return defaultLogger()
}

func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{entry: l.entryOrDefault().WithField(key, value)}
}

func (l *Logger) Debug(args ...any) {
	l.entryOrDefault().Debug(args...)
}

func (l *Logger) Info(args ...any) {
	l.entryOrDefault().Info(args...)
}

func (l *Logger) Warn(args ...any) {
	l.entryOrDefault().Warn(args...)
}

func (l *Logger) Error(args ...any) {
	l.entryOrDefault().Error(args...)
}

func (l *Logger) SetOutput(output io.Writer) {
	if output == nil {
		return
	}

	l.entryOrDefault().Logger.SetOutput(output)
}

func (l *Logger) entryOrDefault() *logrus.Entry {
	if l != nil && l.entry != nil {
		return l.entry
	}

	return defaultLogger().entry
}

func defaultLogger() *Logger {
	log, err := New(Config{})
	if err != nil {
		panic(fmt.Sprintf("init default logger: %v", err))
	}

	return log
}

func (c Config) withDefaults() Config {
	level := strings.TrimSpace(strings.ToLower(c.Level))
	if level == "" {
		level = defaultLevel
	}

	format := strings.TrimSpace(strings.ToLower(c.Format))
	if format == "" {
		format = defaultFormat
	}

	c.Level = level
	c.Format = format
	c.OutputPath = strings.TrimSpace(c.OutputPath)

	return c
}

func buildOutput(outputPath string) (io.Writer, error) {
	if outputPath == "" {
		return os.Stdout, nil
	}

	logDir := filepath.Dir(outputPath)
	if logDir != "." {
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return nil, fmt.Errorf("create log dir %q: %w", logDir, err)
		}
	}

	file, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file %q: %w", outputPath, err)
	}

	return file, nil
}
