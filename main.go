/*
Copyright 2017 Home Office All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name:    "kube-ingress-control",
		Author:  "Rohith Jayawardene",
		Email:   "gambol99@gmail.com",
		Usage:   "is a service used to control which domains a ingress resource is permitted to use",
		Version: fmt.Sprintf("%s (git+sha: %s)", Version, GitSHA),

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "listen",
				Usage:  "the network interace the service should listen on `INTERFACE`",
				Value:  ":8443",
				EnvVar: "LISTEN",
			},
			cli.StringFlag{
				Name:   "tls-cert",
				Usage:  "the path to a file containing the tls certificate `PATH`",
				EnvVar: "TLS_CERT",
			},
			cli.StringFlag{
				Name:   "tls-key",
				Usage:  "the path to a file containing the tls key `PATH`",
				EnvVar: "TLS_KEY",
			},
			cli.StringSliceFlag{
				Name:   "ignore-namespace",
				Usage:  "a collection of namespace you can ignore the policy enforcer",
				EnvVar: "IGNORE_NAMESPACE",
			},
			cli.BoolFlag{
				Name:   "enable-http-logging",
				Usage:  "enable http logging on the service `BOOL`",
				EnvVar: "ENABLE_HTTP_LOGGING",
			},
		},

		Action: func(c *cli.Context) error {
			log.SetFormatter(&log.JSONFormatter{})

			// @step: create the controller
			ctl, err := newController(Config{
				EnableLogging:    c.Bool("enable-logging"),
				IgnoreNamespaces: c.StringSlice("ignore-namespace"),
				Listen:           c.String("listen"),
				TLSCert:          c.String("tls-cert"),
				TLSKey:           c.String("tls-key"),
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "[error] unable to initialize controller, %s", err)
				os.Exit(1)
			}

			// @step: start the service
			if err := ctl.start(); err != nil {
				fmt.Fprintf(os.Stderr, "[error] unable to start controller, %s", err)
				os.Exit(1)
			}

			// step: setup the termination signals
			signalChannel := make(chan os.Signal)
			signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			<-signalChannel

			return nil
		},
	}

	app.Run(os.Args)
}
