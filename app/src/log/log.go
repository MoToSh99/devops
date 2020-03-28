package log

import "github.com/juju/loggo"

const (
	//TraceLevel A constant for defining the logging level to trace
	TraceLevel = loggo.TRACE
	//DebugLevel A constant for defining the logging level to debug
	DebugLevel = loggo.DEBUG
	//InfoLevel A constant for defining the logging level to info
	InfoLevel = loggo.INFO
	//WarningLevel A constant for defining the logging level to warning
	WarningLevel = loggo.WARNING
	//ErrorLevel A constant for defining the logging level to error
	ErrorLevel = loggo.ERROR
	//CriticalLevel A constant for defining the logging level to critical
	CriticalLevel = loggo.CRITICAL
)

var logger = loggo.GetLogger("minitwit")

func logErr(logFunc func(message string, args ...interface{}), str string, err error, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		logFunc(str+": %s", args...)
	}
}

//SetLoggingLevel Sets logger level
func SetLoggingLevel(l loggo.Level) {
	logger.SetLogLevel(l)
}

//InfoErr Logs string on info level if the error is not nil
func InfoErr(str string, err error, args ...interface{}) {
	logErr(logger.Infof, str, err, args)
}

//WarningErr Logs string on warning level if the error is not nil
func WarningErr(str string, err error, args ...interface{}) {
	logErr(logger.Warningf, str, err, args)
}

//ErrorErr Logs string on error level if the error is not nil
func ErrorErr(str string, err error, args ...interface{}) {
	logErr(logger.Errorf, str, err, args)
}

//CriticalErr Logs string on critical level if the error is not nil
func CriticalErr(str string, err error, args ...interface{}) {
	logErr(logger.Criticalf, str, err, args)
}

//DebugErr Logs string on debug level if the error is not nil
func DebugErr(str string, err error, args ...interface{}) {
	logErr(logger.Debugf, str, err, args)
}

//TraceErr Logs string on trace level if the error is not nil
func TraceErr(str string, err error, args ...interface{}) {
	logErr(logger.Tracef, str, err, args)
}

//Info Logs string on info level
func Info(str string, args ...interface{}) {
	logger.Infof(str, args...)
}

//Warning Logs string on warning level
func Warning(str string, args ...interface{}) {
	logger.Warningf(str, args...)
}

//Error Logs string on error level
func Error(str string, args ...interface{}) {
	logger.Errorf(str, args...)
}

//Critical Logs string on critical level
func Critical(str string, args ...interface{}) {
	logger.Criticalf(str, args...)
}

//Debug Logs string on debug level
func Debug(str string, args ...interface{}) {
	logger.Debugf(str, args...)
}

//Trace Logs string on trace level
func Trace(str string, args ...interface{}) {
	logger.Tracef(str, args...)
}
