package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel representa el nivel de log
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String devuelve representación en string del nivel
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger configuración del logger
type Logger struct {
	level       LogLevel
	jsonFormat  bool
	callerInfo  bool
	service     string
	version     string
	environment string
}

// LogEntry entrada de log
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       string                 `json:"level"`
	Service     string                 `json:"service,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Message     string                 `json:"message"`
	Caller      string                 `json:"caller,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// NewLogger crea un nuevo logger
func NewLogger(service string) *Logger {
	return &Logger{
		level:       LevelInfo,
		jsonFormat:  false,
		callerInfo:  true,
		service:     service,
		version:     "1.0.0",
		environment: getEnv("APP_ENV", "development"),
	}
}

// WithLevel configura nivel de log
func (l *Logger) WithLevel(level LogLevel) *Logger {
	l.level = level
	return l
}

// WithJSONFormat configura formato JSON
func (l *Logger) WithJSONFormat(json bool) *Logger {
	l.jsonFormat = json
	return l
}

// WithCallerInfo configura información del llamador
func (l *Logger) WithCallerInfo(caller bool) *Logger {
	l.callerInfo = caller
	return l
}

// WithVersion configura versión
func (l *Logger) WithVersion(version string) *Logger {
	l.version = version
	return l
}

// WithEnvironment configura entorno
func (l *Logger) WithEnvironment(env string) *Logger {
	l.environment = env
	return l
}

// Debug log nivel debug
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	if l.level <= LevelDebug {
		l.log(LevelDebug, msg, fields...)
	}
}

// Info log nivel info
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	if l.level <= LevelInfo {
		l.log(LevelInfo, msg, fields...)
	}
}

// Warn log nivel warn
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	if l.level <= LevelWarn {
		l.log(LevelWarn, msg, fields...)
	}
}

// Error log nivel error
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	if l.level <= LevelError {
		allFields := mergeFields(fields...)
		if err != nil {
			allFields["error"] = err.Error()
		}
		l.log(LevelError, msg, allFields)
	}
}

// Fatal log nivel fatal
func (l *Logger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	if l.level <= LevelFatal {
		allFields := mergeFields(fields...)
		if err != nil {
			allFields["error"] = err.Error()
		}
		l.log(LevelFatal, msg, allFields)
		os.Exit(1)
	}
}

// log escribe el log
func (l *Logger) log(level LogLevel, msg string, fields ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().Format(time.RFC3339),
		Level:       level.String(),
		Service:     l.service,
		Version:     l.version,
		Environment: l.environment,
		Message:     msg,
		Fields:      mergeFields(fields...),
	}

	if l.callerInfo {
		entry.Caller = l.getCallerInfo()
	}

	if l.jsonFormat {
		l.logJSON(entry)
	} else {
		l.logText(entry)
	}
}

// logJSON log en formato JSON
func (l *Logger) logJSON(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("ERROR: failed to marshal log entry: %v", err)
		return
	}

	log.Println(string(data))
}

// logText log en formato texto
func (l *Logger) logText(entry LogEntry) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%s %s", entry.Timestamp, entry.Level))

	if l.service != "" {
		builder.WriteString(fmt.Sprintf(" [%s]", entry.Service))
	}

	if entry.Caller != "" {
		builder.WriteString(fmt.Sprintf(" %s", entry.Caller))
	}

	builder.WriteString(fmt.Sprintf(": %s", entry.Message))

	if len(entry.Fields) > 0 {
		builder.WriteString(" |")
		for key, value := range entry.Fields {
			builder.WriteString(fmt.Sprintf(" %s=%v", key, value))
		}
	}

	if entry.Error != "" {
		builder.WriteString(fmt.Sprintf(" | error=%s", entry.Error))
	}

	log.Println(builder.String())
}

// getCallerInfo obtiene información del llamador
func (l *Logger) getCallerInfo() string {
	// Obtener información 3 niveles arriba (skip: 0=this function, 1=log, 2=Debug/Info/etc.)
	pc := make([]uintptr, 1)
	n := runtime.Callers(4, pc)
	if n == 0 {
		return ""
	}

	frame, _ := runtime.CallersFrames(pc).Next()

	// Extraer solo el nombre del archivo y línea
	file := frame.File
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}

	return fmt.Sprintf("%s:%d", file, frame.Line)
}

// mergeFields combina múltiples mapas de fields
func mergeFields(fields ...map[string]interface{}) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}

	result := make(map[string]interface{})
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			result[key] = value
		}
	}

	return result
}

// MaskSensitiveFields enmascara campos sensibles en los fields
func (l *Logger) MaskSensitiveFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}

	masked := make(map[string]interface{})
	sensitiveKeys := []string{
		"password", "token", "secret", "key",
		"credit_card", "cvv", "ssn", "sin",
		"authorization", "api_key", "private_key",
		"access_token", "refresh_token",
	}

	for key, value := range fields {
		keyLower := strings.ToLower(key)
		maskIt := false

		for _, sensitive := range sensitiveKeys {
			if strings.Contains(keyLower, sensitive) {
				maskIt = true
				break
			}
		}

		if maskIt {
			switch v := value.(type) {
			case string:
				masked[key] = SafeStringForLog(v)
			default:
				masked[key] = "***MASKED***"
			}
		} else {
			masked[key] = value
		}
	}

	return masked
}

// RequestLogger log de requests HTTP
func (l *Logger) RequestLogger(method, path, clientIP string, status int, latency time.Duration, fields ...map[string]interface{}) {
	allFields := mergeFields(fields...)
	allFields["method"] = method
	allFields["path"] = path
	allFields["client_ip"] = SafeStringForLog(clientIP)
	allFields["status"] = status
	allFields["latency"] = latency.String()

	level := LevelInfo
	if status >= 500 {
		level = LevelError
	} else if status >= 400 {
		level = LevelWarn
	}

	l.log(level, "HTTP request", allFields)
}

// DatabaseLogger log de operaciones de base de datos
func (l *Logger) DatabaseLogger(operation, table string, duration time.Duration, rowsAffected int64, err error, fields ...map[string]interface{}) {
	allFields := mergeFields(fields...)
	allFields["operation"] = operation
	allFields["table"] = table
	allFields["duration"] = duration.String()
	allFields["rows_affected"] = rowsAffected

	if err != nil {
		l.Error(fmt.Sprintf("Database %s on %s", operation, table), err, allFields)
	} else {
		l.Info(fmt.Sprintf("Database %s on %s", operation, table), allFields)
	}
}

// BusinessLogger log de operaciones de negocio
func (l *Logger) BusinessLogger(operation, entity string, entityID interface{}, success bool, fields ...map[string]interface{}) {
	allFields := mergeFields(fields...)
	allFields["operation"] = operation
	allFields["entity"] = entity
	allFields["entity_id"] = entityID
	allFields["success"] = success

	level := LevelInfo
	msg := fmt.Sprintf("%s %s", operation, entity)

	if !success {
		level = LevelError
		msg = fmt.Sprintf("Failed to %s %s", strings.ToLower(operation), entity)
	}

	l.log(level, msg, allFields)
}

// PerformanceLogger log de rendimiento
func (l *Logger) PerformanceLogger(operation string, startTime time.Time, threshold time.Duration, fields ...map[string]interface{}) {
	duration := time.Since(startTime)

	allFields := mergeFields(fields...)
	allFields["duration"] = duration.String()
	allFields["duration_ms"] = duration.Milliseconds()

	level := LevelInfo
	if duration > threshold {
		level = LevelWarn
	}

	l.log(level, fmt.Sprintf("Performance: %s", operation), allFields)
}

// AuditLogger log de auditoría
func (l *Logger) AuditLogger(userID, action, resource string, resourceID interface{}, success bool, fields ...map[string]interface{}) {
	allFields := mergeFields(fields...)
	allFields["user_id"] = userID
	allFields["action"] = action
	allFields["resource"] = resource
	allFields["resource_id"] = resourceID
	allFields["success"] = success

	level := LevelInfo
	if !success {
		level = LevelWarn
	}

	l.log(level, fmt.Sprintf("Audit: %s %s", action, resource), allFields)
}

// getEnv obtiene variable de entorno
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GlobalLogger logger global
var GlobalLogger = NewLogger("ticket-system")

// SetGlobalLogger configura logger global
func SetGlobalLogger(logger *Logger) {
	GlobalLogger = logger
}

// LogDebug log global debug
func LogDebug(msg string, fields ...map[string]interface{}) {
	GlobalLogger.Debug(msg, fields...)
}

// LogInfo log global info
func LogInfo(msg string, fields ...map[string]interface{}) {
	GlobalLogger.Info(msg, fields...)
}

// LogWarn log global warn
func LogWarn(msg string, fields ...map[string]interface{}) {
	GlobalLogger.Warn(msg, fields...)
}

// LogError log global error
func LogError(msg string, err error, fields ...map[string]interface{}) {
	GlobalLogger.Error(msg, err, fields...)
}

// LogFatal log global fatal
func LogFatal(msg string, err error, fields ...map[string]interface{}) {
	GlobalLogger.Fatal(msg, err, fields...)
}
