package utils

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
	gormlog "gorm.io/gorm/logger"
)

// #region
type CustomGORMLogger struct {
	Clog log.Helper
}

func (c *CustomGORMLogger) LogMode(gormlog.LogLevel) gormlog.Interface {
	return c
}
func (c *CustomGORMLogger) Info(ctx context.Context, str string, ext ...interface{}) {
	parameter := make([]interface{}, 1)
	parameter = append(parameter, ext...)
	c.Clog.Info(parameter...)
}
func (c *CustomGORMLogger) Warn(ctx context.Context, str string, ext ...interface{}) {
	parameter := make([]interface{}, 1)
	parameter = append(parameter, ext...)
	c.Clog.Warn(parameter...)
}
func (c *CustomGORMLogger) Error(ctx context.Context, str string, ext ...interface{}) {
	parameter := make([]interface{}, 1)
	parameter = append(parameter, ext...)
	c.Clog.Error(parameter...)
}
func (c *CustomGORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// sql, rowsAffected := fc()
	// c.Clog.Debugf("SQL Begin: %v; SQL: %s SQL Rows affected: %d", begin, sql, rowsAffected)
}

// KratosHandler 是一个自定义的 slog.Handler，用于适配 Kratos 的 log.Logger
type KratosHandler struct {
	kratosLogger *log.Helper
}

// NewKratosHandler 创建一个新的 KratosHandler
func NewKratosHandler(logger *log.Helper) *slog.Logger {
	handler := &KratosHandler{
		kratosLogger: logger,
	}
	loggerWithSlog := slog.New(handler)
	return loggerWithSlog
}

// Enabled 检查指定级别的日志是否启用
func (h *KratosHandler) Enabled(_ context.Context, level slog.Level) bool {
	return true
}

// Handle 处理日志记录
func (h *KratosHandler) Handle(_ context.Context, r slog.Record) error {

	h.kratosLogger.Log(log.Level(r.Level), r)
	return nil
}

// WithAttrs 返回一个带有额外属性的新处理程序
func (h *KratosHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup 返回一个带有指定组的新处理程序
func (h *KratosHandler) WithGroup(name string) slog.Handler {
	return h
}

// GinLogger gin logger middleware
func GinLogger(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		log := log.NewHelper(logger)
		// 记录请求日志
		log.Infof("path: %s, query: %s, ip: %s, cost: %v", path, query, c.ClientIP(), cost)
	}
}
