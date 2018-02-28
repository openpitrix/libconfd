// Copyright memkv. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-memkv file.

package libconfd

func _Example() {
	cfg := MustLoadConfig("~/.confd/config.toml")
	client := tNewEnvClient()

	NewProcessor(cfg, client).Run()
}
