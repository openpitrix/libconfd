// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/urfave/cli"

	"openpitrix.io/libconfd"
)

func main() {
	app := cli.NewApp()
	app.Name = "miniconfd"
	app.Usage = "miniconfd is simple confd, only support toml backend file."
	app.Version = "0.1.0"

	app.UsageText = `miniconfd [global options] command [options] [args...]

Example:
   miniconfd list
   miniconfd info
   miniconfd make target
   miniconfd getv key
   miniconfd tour

   miniconfd`

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config",
			Value:  "confd.toml",
			Usage:  "miniconfd config file",
			EnvVar: "MINICONFD_CONFILE_FILE",
		},
	}

	app.Before = func(context *cli.Context) error {
		flag.Parse()
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "list enabled template resource",

			Action: func(c *cli.Context) {
				cfg := libconfd.MustLoadConfig(c.GlobalString("config"))
				client := libconfd.NewFileBackendsClient(cfg.File)

				libconfd.NewApplication(cfg, client).List()
				return
			},
		},
		{
			Name:      "info",
			Usage:     "show template resource info",
			ArgsUsage: "[name...]",

			Action: func(c *cli.Context) {
				cfg := libconfd.MustLoadConfig(c.GlobalString("config"))
				client := libconfd.NewFileBackendsClient(cfg.File)

				libconfd.NewApplication(cfg, client).Info(c.Args().First())
				return
			},
		},

		{
			Name:      "make",
			Usage:     "make template target, not run any command",
			ArgsUsage: "target",

			Action: func(c *cli.Context) {
				cfg := libconfd.MustLoadConfig(c.GlobalString("config"))
				client := libconfd.NewFileBackendsClient(cfg.File)

				libconfd.NewApplication(cfg, client).Make(c.Args().First())
				return
			},
		},

		{
			Name:      "getv",
			Usage:     "get value from backend by key",
			ArgsUsage: "key",

			Action: func(c *cli.Context) {
				cfg := libconfd.MustLoadConfig(c.GlobalString("config"))
				client := libconfd.NewFileBackendsClient(cfg.File)

				libconfd.NewApplication(cfg, client).GetValue(c.Args().First())
				return
			},
		},

		{
			Name:  "tour",
			Usage: "show more examples",
			Action: func(c *cli.Context) {
				fmt.Println(tourTopic)
			},
		},
	}

	app.Action = func(c *cli.Context) {
		if c.NArg() > 0 {
			fmt.Fprintf(c.App.Writer, "invalid: %s; try -h\n", c.Args())
			return
		}
		cfg := libconfd.MustLoadConfig(c.GlobalString("config"))
		client := libconfd.NewFileBackendsClient(cfg.File)

		libconfd.NewApplication(cfg, client).Main()
		return
	}

	app.CommandNotFound = func(ctx *cli.Context, command string) {
		fmt.Fprintf(ctx.App.Writer, "not found '%v'!\n", command)
	}

	app.Run(os.Args)
}

const tourTopic = `
miniconfd list
miniconfd info simple

miniconfd make simple
miniconfd make simple.windows

miniconfd getv /
miniconfd getv /key

miniconfd
`
