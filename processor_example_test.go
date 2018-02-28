// Copyright memkv. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-memkv file.

package libconfd_test

import (
	"github.com/chai2010/libconfd"
	"github.com/chai2010/libconfd/backends/env"
)

func Example() {
	cfg := libconfd.MustLoadConfig("~/.confd/config.toml")
	client := env.NewEnvClient()

	libconfd.NewProcessor(cfg, client).Run()
}
