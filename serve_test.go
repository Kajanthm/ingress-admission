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
	"net/http"
	"testing"

	admission "k8s.io/api/admission/v1alpha1"
	authentication "k8s.io/api/authentication/v1"
	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	fakeHostname = "rohith.dev.homeoffice.gov.uk"
)

func TestIngressNoNamespace(t *testing.T) {
	requests := []request{
		{
			URI:             "/",
			Method:          http.MethodPost,
			AdmissionReview: createFakeIngressReview(fakeHostname),
			ExpectedStatus: &admission.AdmissionReviewStatus{
				Result: &metav1.Status{
					Code:    http.StatusForbidden,
					Message: "unable to get namespace",
					Reason:  metav1.StatusReasonForbidden,
					Status:  metav1.StatusFailure,
				},
			},
			ExpectedCode: http.StatusOK,
		},
	}
	newFakeController().runTests(t, requests)
}

func TestIngressNoAnnotation(t *testing.T) {
	c := newFakeController()
	c.service.client.CoreV1().Namespaces().Create(&api.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	})

	requests := []request{
		{
			URI:             "/",
			Method:          http.MethodPost,
			AdmissionReview: createFakeIngressReview(fakeHostname),
			ExpectedStatus: &admission.AdmissionReviewStatus{
				Result: &metav1.Status{
					Code:    http.StatusForbidden,
					Message: "namespace has no whitelist annotation: ingress-admission.acp.homeoffice.gov.uk/domains",
					Reason:  metav1.StatusReasonForbidden,
					Status:  metav1.StatusFailure,
				},
			},
			ExpectedCode: http.StatusOK,
		},
	}
	c.runTests(t, requests)
}

func TestWhitelistEmpty(t *testing.T) {
	c := newFakeController()
	c.service.client.CoreV1().Namespaces().Create(&api.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Annotations: map[string]string{DomainWhitelistAnnotation: ""},
		},
	})

	requests := []request{
		{
			URI:             "/",
			Method:          http.MethodPost,
			AdmissionReview: createFakeIngressReview(fakeHostname),
			ExpectedStatus: &admission.AdmissionReviewStatus{
				Result: &metav1.Status{
					Code:    http.StatusForbidden,
					Message: "namespace whitelist is empty",
					Reason:  metav1.StatusReasonForbidden,
					Status:  metav1.StatusFailure,
				},
			},
			ExpectedCode: http.StatusOK,
		},
	}
	c.runTests(t, requests)
}

func TestIgnoredNamespace(t *testing.T) {
	c := newFakeController()
	c.service.config.IgnoreNamespaces = []string{"test"}
	c.service.client.CoreV1().Namespaces().Create(&api.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	})
	requests := []request{
		{
			URI:             "/",
			Method:          http.MethodPost,
			AdmissionReview: createFakeIngressReview("rohith.test.svc.cluster.local"),
			ExpectedStatus:  &admission.AdmissionReviewStatus{Allowed: true},
			ExpectedCode:    http.StatusOK,
		},
	}
	c.runTests(t, requests)
}

func TestIgnoredNamespaceBad(t *testing.T) {
	c := newFakeController()
	c.service.config.IgnoreNamespaces = []string{"other_namespae"}
	c.service.client.CoreV1().Namespaces().Create(&api.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	})
	requests := []request{
		{
			URI:             "/",
			Method:          http.MethodPost,
			AdmissionReview: createFakeIngressReview("rohith.test.svc.cluster.local"),
			ExpectedStatus: &admission.AdmissionReviewStatus{
				Result: &metav1.Status{
					Code:    http.StatusForbidden,
					Message: "namespace has no whitelist annotation: ingress-admission.acp.homeoffice.gov.uk/domains",
					Reason:  metav1.StatusReasonForbidden,
					Status:  metav1.StatusFailure,
				},
			},
			ExpectedCode: http.StatusOK,
		},
	}
	c.runTests(t, requests)
}

func TestNamespaceWhitelist(t *testing.T) {
	c := newFakeController()
	c.service.client.CoreV1().Namespaces().Create(&api.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Annotations: map[string]string{DomainWhitelistAnnotation: "*.test.svc.cluster.local"},
		},
	})
	requests := []request{
		{
			URI:             "/",
			Method:          http.MethodPost,
			AdmissionReview: createFakeIngressReview("rohith.test.svc.cluster.local"),
			ExpectedStatus:  &admission.AdmissionReviewStatus{Allowed: true},
			ExpectedCode:    http.StatusOK,
		},
		{
			URI:             "/",
			Method:          http.MethodPost,
			AdmissionReview: createFakeIngressReview("site.test.svc.cluster.local"),
			ExpectedStatus:  &admission.AdmissionReviewStatus{Allowed: true},
			ExpectedCode:    http.StatusOK,
		},
		{
			URI:             "/",
			Method:          http.MethodPost,
			AdmissionReview: createFakeIngressReview("bad.test.test.svc.cluster.local"),
			ExpectedStatus: &admission.AdmissionReviewStatus{
				Result: &metav1.Status{
					Code:    http.StatusForbidden,
					Message: "hostname: bad.test.test.svc.cluster.local is not permitted by namespace policy",
					Reason:  metav1.StatusReasonForbidden,
					Status:  metav1.StatusFailure,
				},
			},
			ExpectedCode: http.StatusOK,
		},
	}
	c.runTests(t, requests)
}

func TestVersionHandler(t *testing.T) {
	requests := []request{
		{
			URI:             "/version",
			ExpectedCode:    http.StatusOK,
			ExpectedContent: Version + "\n",
		},
	}
	newFakeController().runTests(t, requests)
}

func TestHealthHandler(t *testing.T) {
	requests := []request{
		{
			URI:          "/health",
			ExpectedCode: http.StatusOK,
		},
	}
	newFakeController().runTests(t, requests)
}

func createFakeIngress(hostname string) *extensions.Ingress {
	if hostname == "" {
		hostname = "rohith.dev.homeoffice.gov.uk"
	}
	return &extensions.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: extensions.IngressSpec{
			TLS: []extensions.IngressTLS{
				{
					Hosts:      []string{hostname},
					SecretName: "tls",
				},
			},
			Rules: []extensions.IngressRule{
				{
					Host: hostname,
				},
			},
		},
		Status: extensions.IngressStatus{
			LoadBalancer: api.LoadBalancerStatus{
				Ingress: []api.LoadBalancerIngress{
					{
						IP:       "",
						Hostname: "",
					},
				},
			},
		},
	}
}

func createFakeIngressReview(hostname string) *admission.AdmissionReview {
	ingress := createFakeIngress(hostname)
	// we need to encode the ingress
	content, _ := json.Marshal(ingress)

	return &admission.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1alpha1",
		},
		Spec: admission.AdmissionReviewSpec{
			Kind: metav1.GroupVersionKind{
				Group:   "extensions",
				Version: "v1beta1",
				Kind:    "Ingress",
			},
			Object:    runtime.RawExtension{Raw: content},
			Operation: admission.Create,
			Name:      "test",
			Namespace: "test",
			Resource: metav1.GroupVersionResource{
				Group:    "extensions",
				Version:  "v1beta1",
				Resource: "ingresses",
			},
			UserInfo: authentication.UserInfo{
				Username: "admin",
			},
		},
	}
}
