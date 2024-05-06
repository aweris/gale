package log

// logger is the default global logger.
var logger = NewLogger()

// StartGroup starts a new group in the default logger.
func StartGroup() {
	logger.StartGroup()
}

// EndGroup ends the current group in the default logger.
func EndGroup() {
	logger.EndGroup()
}

// Info logs an info message in the default logger.
func Info(message string) {
	logger.Info(message)
}

// Infof logs an info message in the default logger.
func Infof(message string, keyvals ...interface{}) {
	logger.Infof(message, keyvals...)
}

// Debug logs a debug message in the default logger.
func Debug(message string) {
	logger.Debug(message)
}

// Debugf logs a debug message in the default logger.
func Debugf(message string, keyvals ...interface{}) {
	logger.Debugf(message, keyvals...)
}

// Warn logs a warning message in the default logger.
func Warn(message string) {
	logger.Warn(message)
}

// Warnf logs a warning message in the default logger.
func Warnf(message string, keyvals ...interface{}) {
	logger.Warnf(message, keyvals...)
}

// Error logs an error message in the default logger.
func Error(message string) {
	logger.Error(message)
}

// Errorf logs an error message in the default logger.
func Errorf(message string, keyvals ...interface{}) {
	logger.Errorf(message, keyvals...)
}

// Notice logs a notice message in the default logger.
func Notice(message string) {
	logger.Notice(message)
}

// Noticef logs a notice message in the default logger.
func Noticef(message string, keyvals ...interface{}) {
	logger.Noticef(message, keyvals...)
}
