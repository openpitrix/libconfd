// Copyright memkv. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-memkv file.

package libconfd_test

import (
	"context"

	"github.com/chai2010/libconfd"
)

func Example() {
	confd := libconfd.NewProcessor(
		libconfd.MustLoadConfig("~/.confd/config.toml"),
		libconfd.NewEnvBackendClient(),
	)
	confd.Run(context.Background())
}
