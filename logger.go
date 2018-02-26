// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"fmt"
	"io"
	"os"

	"github.com/golang/glog"
)

var logger Logger = NewStdLogger(os.Stderr)

// NewStdLogger create new logger based on std log.
// If defaultLevel missing, use WARNING as the default level.
// Level: DEBUG < INFO < WARNING < ERROR < PANIC < FATAL
func NewStdLogger(out io.Writer, defaultLevel ...string) Logger {
	return new(glogger)
}

func GetLogger() Logger {
	return logger
}

func SetLogger(new Logger) (old Logger) {
	old, logger = logger, new
	return
}

type Logger interface {
	Debug(args ...interface{})
	Debugln(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Warning(args ...interface{})
	Warningln(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicln(args ...interface{})
	Panicf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})

	// Level: DEBUG < INFO < WARNING < ERROR < PANIC < FATAL
	GetLevel() string
	SetLevel(new string) (old string)

	// V reports whether verbosity level l is at least the requested verbose level.
	//V(l int) bool
}

type glogger struct{}

func (_ *glogger) Debug(args ...interface{})                 {}
func (_ *glogger) Debugln(args ...interface{})               {}
func (_ *glogger) Debugf(format string, args ...interface{}) {}

func (_ *glogger) Panic(args ...interface{})                 {}
func (_ *glogger) Panicln(args ...interface{})               {}
func (_ *glogger) Panicf(format string, args ...interface{}) {}

func (_ *glogger) GetLevel() string                 { return "" }
func (_ *glogger) SetLevel(new string) (old string) { return "" }

func (_ *glogger) Info(args ...interface{}) {
	glog.InfoDepth(1, args...)
}

func (_ *glogger) Infoln(args ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintln(args...))
}

func (_ *glogger) Infof(format string, args ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintf(format, args...))
}

func (_ *glogger) Warning(args ...interface{}) {
	glog.WarningDepth(1, args...)
}

func (_ *glogger) Warningln(args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintln(args...))
}

func (_ *glogger) Warningf(format string, args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintf(format, args...))
}

func (_ *glogger) Error(args ...interface{}) {
	glog.ErrorDepth(1, args...)
}

func (_ *glogger) Errorln(args ...interface{}) {
	glog.ErrorDepth(1, fmt.Sprintln(args...))
}

func (_ *glogger) Errorf(format string, args ...interface{}) {
	glog.ErrorDepth(1, fmt.Sprintf(format, args...))
}

func (_ *glogger) Fatal(args ...interface{}) {
	glog.FatalDepth(1, args...)
}

func (_ *glogger) Fatalln(args ...interface{}) {
	glog.FatalDepth(1, fmt.Sprintln(args...))
}

func (_ *glogger) Fatalf(format string, args ...interface{}) {
	glog.FatalDepth(1, fmt.Sprintf(format, args...))
}

func (_ *glogger) V(l int) bool {
	return bool(glog.V(glog.Level(l)))
}
