// Copyright memkv. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-memkv file.

package libconfd_test

import (
	"github.com/chai2010/libconfd"
)

func Example() {
	cfg := libconfd.MustLoadConfig("~/.confd/config.toml")
	client := libconfd.NewEnvClient()

	libconfd.NewProcessor(cfg, client).Run()
}
