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
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// controller
type controller struct {
	// client is the kubernetes client
	client *kubernetes.Clientset
	// engine the http server
	engine *echo.Echo
	// config is the configuration for the service
	config *Config
}

// newController creates, registers and starts the admission controller
func newController(cfg Config) (*controller, error) {
	log.Infof("starting the ingress admission controller, version: %s, listen: %s", Version, cfg.Listen)
	c := &controller{
		config: &cfg,
	}
	// @step: create the http service
	c.engine = echo.New()
	c.engine.Use(middleware.Recover())
	if cfg.EnableLogging {
		c.engine.Use(middleware.Logger())
	}
	c.engine.GET("/review", c.reviewHandler)
	c.engine.GET("/health", c.healthHandler)
	c.engine.GET("/version", c.versionHandler)

	return c, nil
}

// run start's the http service
func (c *controller) run() error {
	// @step: attempt to create a kubernetes client
	client, err := getKubernetesClient()
	if err != nil {
		return err
	}
	c.client = client

	// @step: configure the http server
	tlsConfig, err := buildTLSConfig(c.config)
	if err != nil {
		return err
	}

	// @step: create the http service
	hs := &http.Server{
		Addr:         c.config.Listen,
		Handler:      c.engine,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
		TLSConfig:    tlsConfig,
	}

	// @step: start the http service
	go func() {
		if err := c.engine.StartServer(hs); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatal("unable to create the http service")
		}
	}()

	return nil
}

// buildTLSConfig builds the TLS configuration from the options
func buildTLSConfig(cfg *Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	// @step: load the certificates
	cert, err := tls.LoadX509KeyPair(cfg.TLSCert, cfg.TLSKey)
	if err != nil {
		return nil, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// @step: are we using client mutual tls?
	if cfg.EnableClientTLS {
		clientCA, err := ioutil.ReadFile(cfg.TLSCA)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(clientCA)
		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

// getKubernetesClient returns a kubernetes api client for us
func getKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
