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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	log "github.com/sirupsen/logrus"
	admission "k8s.io/api/admission/v1alpha1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type controller struct {
	client kubernetes.Interface
	engine *echo.Echo
	config *Config
}

// newController creates, registers and starts the admission controller
func newController(cfg Config) (*controller, error) {
	log.Infof("starting the ingress admission controller, version: %s, listen: %s", Version, cfg.Listen)
	c := &controller{config: &cfg}

	c.engine = echo.New()
	c.engine.HideBanner = true
	c.engine.Use(middleware.Recover())
	if cfg.EnableLogging {
		c.engine.Use(middleware.Logger())
	}
	c.engine.POST("/", c.reviewHandler)
	c.engine.GET("/health", c.healthHandler)
	c.engine.GET("/version", c.versionHandler)

	return c, nil
}

// admit is responsible for applying the policy on the incoming request
func (c *controller) admit(review *admission.AdmissionReview) error {

	ok, message := func() (bool, string) {
		// @check if the object is a ingress
		kind := review.Spec.Kind.Kind
		if kind != "Ingress" {
			return false, fmt.Sprintf("invalid object for review: %s, expected: ingress", kind)
		}

		ingress := &extensions.Ingress{}
		if err := json.Unmarshal(review.Spec.Object.Raw, ingress); err != nil {
			return false, fmt.Sprintf("unable to decode ingress spec: %s", err)
		}

		// @check if this namesapce is being ignored
		for _, x := range c.config.IgnoreNamespaces {
			if x == review.Spec.Namespace {
				log.WithFields(log.Fields{
					"name":      review.Spec.Name,
					"namespace": review.Spec.Namespace,
				}).Info("ignoring the policy enforcement on this namespace")

				return true, ""
			}
		}

		// @check the domain being requested it whitelisted on the namespace
		namespace, err := c.client.CoreV1().Namespaces().Get(review.Spec.Namespace, metav1.GetOptions{})
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err.Error(),
				"namespace": review.Spec.Namespace,
			}).Error("unable to retrieve namespace")

			return false, "unable to get namespace"
		}

		// @check the annotation exists on the namespace
		whitelist, found := namespace.GetAnnotations()[DomainWhitelistAnnotation]
		if !found {
			return false, fmt.Sprintf("namespace has no whitelist annotation: %s", DomainWhitelistAnnotation)
		}

		// @check the whitelist is not empty
		if whitelist == "" {
			return false, "namespace whitelist is empty"
		}
		whitelistedDomains := strings.Split(whitelist, ",")

		// @check if the hostname is covered by the whitelist
		for _, rule := range ingress.Spec.Rules {
			if found := hasDomain(rule.Host, whitelistedDomains); !found {
				return false, fmt.Sprintf("hostname: %s is not permitted by namespace policy", rule.Host)
			}
		}

		return true, ""
	}()
	if !ok {
		log.WithFields(log.Fields{
			"namespace": review.Spec.Namespace,
			"error":     message,
		}).Warn(message)

		review.Status.Allowed = false
		review.Status.Result = &metav1.Status{
			Code:    http.StatusForbidden,
			Message: message,
			Reason:  metav1.StatusReasonForbidden,
			Status:  metav1.StatusFailure,
		}

		return nil
	}

	review.Status.Allowed = true

	return nil
}

// start is repsonsible for starting the service up
func (c *controller) start() error {
	// @step: attempt to create a kubernetes client
	client, err := getKubernetesClient()
	if err != nil {
		return err
	}
	c.client = client

	// @step: configure the http server
	tlsConfig, err := getTLSConfig(c.config)
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
