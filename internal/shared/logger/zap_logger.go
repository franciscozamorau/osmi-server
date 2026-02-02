package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger implementa el logger con zap
type ZapLogger struct {
	*zap.SugaredLogger
	zapLogger *zap.Logger
}

// NewZapLogger crea un nuevo logger zap
func NewZapLogger(environment string) *ZapLogger {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.LevelKey = "level"
		config.EncoderConfig.CallerKey = "caller"
		config.DisableStacktrace = false
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	// Configurar salida
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Construir logger
	zapLogger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		panic("failed to create zap logger: " + err.Error())
	}

	return &ZapLogger{
		SugaredLogger: zapLogger.Sugar(),
		zapLogger:     zapLogger,
	}
}

// Sync sincroniza el logger
func (l *ZapLogger) Sync() error {
	return l.zapLogger.Sync()
}

// WithFields crea un logger con campos adicionales
func (l *ZapLogger) WithFields(fields map[string]interface{}) *ZapLogger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	newLogger := l.zapLogger.With(zapFields...)
	return &ZapLogger{
		SugaredLogger: newLogger.Sugar(),
		zapLogger:     newLogger,
	}
}

// WithError crea un logger con un error
func (l *ZapLogger) WithError(err error) *ZapLogger {
	return l.WithFields(map[string]interface{}{
		"error": err.Error(),
	})
}

// WithRequest crea un logger con información de la petición
func (l *ZapLogger) WithRequest(method, path, ip string) *ZapLogger {
	return l.WithFields(map[string]interface{}{
		"method": method,
		"path":   path,
		"ip":     ip,
	})
}

// WithUser crea un logger con información del usuario
func (l *ZapLogger) WithUser(userID, email string) *ZapLogger {
	return l.WithFields(map[string]interface{}{
		"user_id": userID,
		"email":   email,
	})
}

// DatabaseLogger log para operaciones de base de datos
func (l *ZapLogger) DatabaseLogger(operation, table string, duration time.Duration, rowsAffected int64, err error, data map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":     operation,
		"table":         table,
		"duration_ms":   duration.Milliseconds(),
		"rows_affected": rowsAffected,
		"type":          "database",
	}

	for k, v := range data {
		fields[k] = v
	}

	if err != nil {
		fields["error"] = err.Error()
		l.Errorw("Database operation failed", fields)
	} else {
		l.Infow("Database operation completed", fields)
	}
}

// APILogger log para operaciones de API
func (l *ZapLogger) APILogger(method, endpoint string, statusCode int, duration time.Duration, err error, data map[string]interface{}) {
	fields := map[string]interface{}{
		"method":      method,
		"endpoint":    endpoint,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"type":        "api",
	}

	for k, v := range data {
		fields[k] = v
	}

	if err != nil {
		fields["error"] = err.Error()
		l.Errorw("API request failed", fields)
	} else {
		l.Infow("API request completed", fields)
	}
}

// StructuredLogger interface para logging estructurado
type StructuredLogger interface {
	Info(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Debug(msg string, fields map[string]interface{})
	Fatal(msg string, fields map[string]interface{})
}

// Global logger instance
var (
	globalLogger *ZapLogger
)

// InitGlobalLogger inicializa el logger global
func InitGlobalLogger(environment string) {
	globalLogger = NewZapLogger(environment)
}

// GetLogger retorna el logger global
func GetLogger() *ZapLogger {
	if globalLogger == nil {
		// Logger por defecto para desarrollo
		globalLogger = NewZapLogger("development")
	}
	return globalLogger
}

// Info log a nivel info
func Info(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		InitGlobalLogger("development")
	}
	globalLogger.zapLogger.Info(msg, fields...)
}

// Error log a nivel error
func Error(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		InitGlobalLogger("development")
	}
	globalLogger.zapLogger.Error(msg, fields...)
}

// Warn log a nivel warn
func Warn(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		InitGlobalLogger("development")
	}
	globalLogger.zapLogger.Warn(msg, fields...)
}

// Debug log a nivel debug
func Debug(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		InitGlobalLogger("development")
	}
	globalLogger.zapLogger.Debug(msg, fields...)
}

// Fatal log a nivel fatal
func Fatal(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		InitGlobalLogger("development")
	}
	globalLogger.zapLogger.Fatal(msg, fields...)
}

// SyncGlobal sincroniza el logger global
func SyncGlobal() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// Helper para crear campos rápido
func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// RequestIDField crea un campo para request ID
func RequestIDField(requestID string) zap.Field {
	return zap.String("request_id", requestID)
}

// UserIDField crea un campo para user ID
func UserIDField(userID string) zap.Field {
	return zap.String("user_id", userID)
}

// DurationField crea un campo para duración
func DurationField(duration time.Duration) zap.Field {
	return zap.Duration("duration", duration)
}

// FileLogger crea un logger que escribe a archivo
func FileLogger(logPath string) *ZapLogger {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{logPath}
	config.ErrorOutputPaths = []string{logPath}

	zapLogger, err := config.Build()
	if err != nil {
		panic("failed to create file logger: " + err.Error())
	}

	return &ZapLogger{
		SugaredLogger: zapLogger.Sugar(),
		zapLogger:     zapLogger,
	}
}

// MultiLogger crea un logger que escribe a múltiples destinos
func MultiLogger(destinations ...string) *ZapLogger {
	config := zap.NewProductionConfig()
	config.OutputPaths = destinations
	config.ErrorOutputPaths = destinations

	zapLogger, err := config.Build()
	if err != nil {
		// Fallback a stdout
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
		zapLogger, _ = config.Build()
	}

	return &ZapLogger{
		SugaredLogger: zapLogger.Sugar(),
		zapLogger:     zapLogger,
	}
}

// TestLogger crea un logger para testing
func TestLogger() *ZapLogger {
	if os.Getenv("TEST_LOGGER") == "true" {
		return NewZapLogger("development")
	}

	// Logger silencioso para tests
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	config.DisableCaller = true
	config.DisableStacktrace = true

	zapLogger, _ := config.Build()
	return &ZapLogger{
		SugaredLogger: zapLogger.Sugar(),
		zapLogger:     zapLogger,
	}
}
