// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

func ExampleLogger() {
	var logger Logger = NewGlogger()

	logger.Info("hello")
	logger.Warning("confd")

	if logger.V(1) {
		logger.Info("debug: ...")
	}
}
