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
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/admission"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

// admit is responsible for applying the policy on the incoming request
func (c *controller) admit(review *admission.AdmissionReview) (admission.AdmissionReviewStatus, error) {
	var status admission.AdmissionReviewStatus

	// @step: check the domain being requested it whitelisted on the namespace
	namespace, err := c.client.CoreV1().Namespaces().Get(review.Spec.Namespace, metav1.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"namespace": review.Spec.Namespace,
		}).Error("unable to retrieve namespace")

		status.Result.Message = "unable to get the namespace"

		return status, nil
	}

	// @step: extract the ingress spec from the resource object
	var ingress *extensions.IngressSpec

	ok, message := func() (bool, string) {
		// @check the annotation exists on the namespace
		whitelist, found := namespace.GetAnnotations()[DomainWhitelistAnnotation]
		if !found {
			return false, "namespace has no whitelist annotation"
		}
		// @check the whitelist is not empty
		if whitelist == "" {
			return false, "namespace whitelist is empty"
		}
		// @check if the hostname is covered by the whitelist
		whitelistedDomains := strings.Split(whitelist, ",")
		for _, rule := range ingress.Rules {
			if found := hasDomain(rule.Host, whitelistedDomains); !found {
				return false, fmt.Sprintf("hostname %s is not permitted by namespace", rule.Host)
			}
		}

		return true, ""
	}()
	if !ok {
		log.WithFields(log.Fields{
			"namespace": review.Spec.Namespace,
		}).Warn(message)

		status.Result.Message = message

		return status, nil
	}
	status.Allowed = true

	return status, nil
}

// hasDomain checks the domain exists with in the whitelist
// e.g hostname.namespace.svc.cluster.local or *.namespace.svc.cluster.local
func hasDomain(hostname string, whitelist []string) bool {
	for _, domain := range whitelist {
		wildcard := strings.HasPrefix(domain, "*.")
		switch wildcard {
		case true:
			// a quick hacky check to ensure the you don't have subdomains
			size := len(strings.Split(domain, "."))
			hostSize := len(strings.Split(hostname, "."))
			if size != hostSize {
				return false
			}

			domain = strings.TrimPrefix(domain, "*")
			if strings.HasSuffix(hostname, domain) {
				return true
			}
		default:
			// @check there is an exact match between hostname and whitelist
			if hostname == domain {
				return true
			}
		}
	}

	return false
}
