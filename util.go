// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"os"
	"path"
	"path/filepath"
)

// fileInfo describes a configuration file and is returned by fileStat.
type fileInfo struct {
	Uid  uint32
	Gid  uint32
	Mode os.FileMode
	Md5  string
}

func utilAppendPrefix(prefix string, keys []string) []string {
	s := make([]string, len(keys))
	for i, k := range keys {
		s[i] = path.Join(prefix, k)
	}
	return s
}

// utilFileExist reports whether path exits.
func utilFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}

// utilSameConfig reports whether src and dest config files are equal.
// Two config files are equal when they have the same file contents and
// Unix permissions. The owner, group, and mode must match.
// It return false in other cases.
func utilSameConfig(src, dest string) (bool, error) {
	if !utilFileExist(dest) {
		return false, nil
	}

	d, err := utilFileStat(dest)
	if err != nil {
		return false, err
	}
	s, err := utilFileStat(src)
	if err != nil {
		return false, err
	}

	if d.Uid != s.Uid {
		logger.Infof("%s has UID %d should be %d", dest, d.Uid, s.Uid)
	}
	if d.Gid != s.Gid {
		logger.Infof("%s has GID %d should be %d", dest, d.Gid, s.Gid)
	}
	if d.Mode != s.Mode {
		logger.Infof("%s has mode %s should be %s", dest, os.FileMode(d.Mode), os.FileMode(s.Mode))
	}
	if d.Md5 != s.Md5 {
		logger.Infof("%s has md5sum %s should be %s", dest, d.Md5, s.Md5)
	}

	if d.Uid != s.Uid || d.Gid != s.Gid || d.Mode != s.Mode || d.Md5 != s.Md5 {
		return false, nil
	}

	return true, nil
}

// utilRecursiveFindFiles find files with pattern in the root with depth.
func utilRecursiveFindFiles(root string, pattern string) (files []string, err error) {
	err = filepath.Walk(root, func(path string, f os.FileInfo, err error) (inner error) {
		if err != nil || f.IsDir() {
			return
		}
		if matched, _ := filepath.Match(pattern, f.Name()); matched {
			files = append(files, path)
		}
		return
	})
	return
}
