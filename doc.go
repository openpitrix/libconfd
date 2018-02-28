// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

/*
Package libconfd provides mini confd lib.

Examples

Build a simple confd:

	package main

	import (
		"github.com/chai2010/libconfd"
		"github.com/chai2010/libconfd/backends/etcd"
	)

	func main() {
		cfg := libconfd.MustLoadConfig("~/.confd/config.toml")
		client := etcd.NewEtcdClient()

		libconfd.NewProcessor(cfg, client).Run()
	}

BUGS

Report bugs to <chaishushan@gmail.com>.
Thanks!
*/
package libconfd
