package log

import "github.com/juju/loggo"

var logger = loggo.GetLogger("minitwit")

//Info Logs string on info level
func Info(str string, err error, args ...interface{}) {
	args = append(args, err)
	logger.Infof(str+": %s", args...)
}

//Warning Logs string on warning level
func Warning(str string, err error, args ...interface{}) {
	args = append(args, err)
	logger.Warningf(str+": %s", args...)
}

//Error Logs string on error level
func Error(str string, err error, args ...interface{}) {
	args = append(args, err)
	logger.Errorf(str+": %s", args...)
}

//Critical Logs string on critical level
func Critical(str string, err error, args ...interface{}) {
	args = append(args, err)
	logger.Criticalf(str+": %s", args...)
}

//Debug Logs string on debug level
func Debug(str string, err error, args ...interface{}) {
	args = append(args, err)
	logger.Debugf(str+": %s", args...)
}

//Trace Logs string on trace level
func Trace(str string, err error, args ...interface{}) {
	args = append(args, err)
	logger.Tracef(str+": %s", args...)
}
