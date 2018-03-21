// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package libconfd_test

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"openpitrix.io/libconfd"
)

func Example() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient("./confd-backend.toml")

	libconfd.NewProcessor().Run(cfg, client)
}

func Example_async() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient("./confd-backend.toml")

	call := libconfd.NewProcessor().Go(cfg, client)

	// do some thing

	call = <-call.Done // will be equal to call
	fmt.Println(call.Error)
}

func Example_multiSync() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient("./confd-backend.toml")

	go libconfd.NewProcessor().Run(cfg, client)
	go libconfd.NewProcessor().Run(cfg, client)

	<-make(chan bool)
}

func Example_multiAsync() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient("./confd-backend.toml")

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
	client := libconfd.NewFileBackendsClient("./confd-backend.toml")

	libconfd.NewProcessor().Run(cfg, client,
		libconfd.WithIntervalMode(),
	)
}
func Example_close() {
	cfg := libconfd.MustLoadConfig("./confd.toml")
	client := libconfd.NewFileBackendsClient("./confd-backend.toml")

	p := libconfd.NewProcessor()

	go func() {
		defer p.Close()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
	}()

	p.Run(cfg, client)
}

func Example_logger() {
	var logger = libconfd.GetLogger()

	logger.SetLevel("DEBUG")
	logger.Debug("1+1=2")
	logger.Info("hello")
}
