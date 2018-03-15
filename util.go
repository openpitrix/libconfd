// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// fileInfo describes a configuration file and is returned by readFileStat.
type fileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode os.FileMode
	Md5  string
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !fi.IsDir() {
		return false
	}
	return true
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func fileNotExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return false
}

// findFilesRecursive find files with pattern in the rootdir with depth.
func findFilesRecursive(rootdir, pattern string) (files []string, err error) {
	err = filepath.Walk(rootdir, func(path string, f os.FileInfo, err error) (inner error) {
		if err != nil || f.IsDir() {
			return
		}
		if matched, _ := filepath.Match(pattern, f.Name()); matched {
			files = append(files, path)
		}
		return
	})
	sort.Strings(files)
	return
}

func getFuncName(skips ...int) string {
	var skip = 1
	if len(skips) > 0 {
		skip = skips[0]
	}

	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	name := runtime.FuncForPC(pc).Name()
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

func strInStrList(s string, ss []string) bool {
	for _, t := range ss {
		if s == t {
			return true
		}
	}
	return false
}
