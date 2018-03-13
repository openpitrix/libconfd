// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/chai2010/libconfd"
)

func Example() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient(cfg.File)

	libconfd.NewProcessor().Run(cfg, client)
}

func Example_async() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient(cfg.File)

	call := libconfd.NewProcessor().Go(cfg, client)

	// do some thing

	call = <-call.Done // will be equal to call
	fmt.Println(call.Error)
}

func Example_multiSync() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient(cfg.File)

	go libconfd.NewProcessor().Run(cfg, client)
	go libconfd.NewProcessor().Run(cfg, client)

	<-make(chan bool)
}

func Example_multiAsync() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient(cfg.File)

	var callList = []*libconfd.Call{
		libconfd.NewProcessor().Go(cfg, client),
		libconfd.NewProcessor().Go(cfg, client),
	}

	var wg sync.WaitGroup
	for i := 0; i < len(callList); i++ {
		wg.Add(1)
		go func(i int) {
			<-callList[i].Done
			wg.Done()
		}(i)
	}
	wg.Wait()

	fmt.Println("Done")
}

func Example_option() {

	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient(cfg.File)

	libconfd.NewProcessor().Run(cfg, client,
		libconfd.WithInterval(time.Second*10),
		libconfd.WithIntervalMode(),
	)
}
