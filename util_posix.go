// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

// +build !windows

package libconfd

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
)

// utilFileStat return a fileInfo describing the named file.
func utilFileStat(name string) (fi fileInfo, err error) {
	if !utilFileExist(name) {
		err = errors.New("File not found")
		return
	}

	f, err := os.Open(name)
	if err != nil {
		return fi, err
	}
	defer f.Close()

	stats, err := f.Stat()
	if err != nil {
		return
	}

	fi.Uid = stats.Sys().(*syscall.Stat_t).Uid
	fi.Gid = stats.Sys().(*syscall.Stat_t).Gid
	fi.Mode = stats.Mode()

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return
	}

	fi.Md5 = fmt.Sprintf("%x", h.Sum(nil))
	return fi, nil
}
