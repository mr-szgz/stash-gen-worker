package logger

import "log"

func Trace(args ...interface{})                 { log.Print(args...) }
func Tracef(format string, args ...interface{}) { log.Printf(format, args...) }
func Debug(args ...interface{})                 { log.Print(args...) }
func Debugf(format string, args ...interface{}) { log.Printf(format, args...) }
func Info(args ...interface{})                  { log.Print(args...) }
func Infof(format string, args ...interface{})  { log.Printf(format, args...) }
func Warn(args ...interface{})                  { log.Print(args...) }
func Warnf(format string, args ...interface{})  { log.Printf(format, args...) }
func Error(args ...interface{})                 { log.Print(args...) }
func Errorf(format string, args ...interface{}) { log.Printf(format, args...) }
func Fatal(args ...interface{})                 { log.Fatal(args...) }
func Fatalf(format string, args ...interface{}) { log.Fatalf(format, args...) }
