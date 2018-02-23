// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"fmt"

	"github.com/golang/glog"
)

type Logger interface {
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Warning(args ...interface{})
	Warningln(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})

	// V reports whether verbosity level l is at least the requested verbose level.
	V(l int) bool
}

type Glogger struct{}

var logger Logger = new(Glogger)

func SetLogger(l Logger) {
	logger = l
}

func (g *Glogger) Info(args ...interface{}) {
	glog.InfoDepth(1, args...)
}

func (g *Glogger) Infoln(args ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintln(args...))
}

func (g *Glogger) Infof(format string, args ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintf(format, args...))
}

func (g *Glogger) Warning(args ...interface{}) {
	glog.WarningDepth(1, args...)
}

func (g *Glogger) Warningln(args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintln(args...))
}

func (g *Glogger) Warningf(format string, args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintf(format, args...))
}

func (g *Glogger) Error(args ...interface{}) {
	glog.ErrorDepth(1, args...)
}

func (g *Glogger) Errorln(args ...interface{}) {
	glog.ErrorDepth(1, fmt.Sprintln(args...))
}

func (g *Glogger) Errorf(format string, args ...interface{}) {
	glog.ErrorDepth(1, fmt.Sprintf(format, args...))
}

func (g *Glogger) Fatal(args ...interface{}) {
	glog.FatalDepth(1, args...)
}

func (g *Glogger) Fatalln(args ...interface{}) {
	glog.FatalDepth(1, fmt.Sprintln(args...))
}

func (g *Glogger) Fatalf(format string, args ...interface{}) {
	glog.FatalDepth(1, fmt.Sprintf(format, args...))
}

func (g *Glogger) V(l int) bool {
	return bool(glog.V(glog.Level(l)))
}
