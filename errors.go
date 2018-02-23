// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotExist = errors.New("libconfd: key does not exist")
	ErrNoMatch  = errors.New("libconfd: no keys match")
)

type KeyError struct {
	Key string
	Err error
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("%v: %s", e.Err, e.Key)
}

func notDeviceOrResourceBusyError(err error) bool {
	if err != nil && strings.Contains(err.Error(), "device or resource busy") {
		return true // OK
	}
	return false
}
