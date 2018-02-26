// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"os"
)

func ExampleLogger() {
	var logger Logger = NewStdLogger(os.Stderr)

	logger.Debug("debug: ...")
	logger.Info("hello")
	logger.Warning("confd")
}
