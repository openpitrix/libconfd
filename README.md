# libconfd

[![Build Status](https://travis-ci.org/chai2010/libconfd.svg)](https://travis-ci.org/chai2010/libconfd)
[![Go Report Card](https://goreportcard.com/badge/github.com/chai2010/libconfd)](https://goreportcard.com/report/github.com/chai2010/libconfd)
[![GoDoc](https://godoc.org/github.com/chai2010/libconfd?status.svg)](https://godoc.org/github.com/chai2010/libconfd)
[![License](http://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/chai2010/libconfd/blob/master/LICENSE)

mini confd lib, based on [confd](https://github.com/kelseyhightower/confd)/[memkv](https://github.com/kelseyhightower/memkv)/[secconf](https://github.com/xordataexchange/crypt).


```go
package main

import (
	"github.com/chai2010/libconfd"
)

func main() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient(cfg.File)

	libconfd.NewProcessor().Run(cfg, client)
}
```
