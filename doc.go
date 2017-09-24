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
	admission "k8s.io/api/admission/v1alpha1"
	authentication "k8s.io/api/authentication/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AdmissionControllerName is the name we register as
	AdmissionControllerName = "ingress-admission.acp.homeoffice.gov.uk"
	// DomainWhitelistAnnotation is the annotation which controls which domains you can use
	DomainWhitelistAnnotation = "ingress-admission.acp.homeoffice.gov.uk/domains"
)

var (
	// Version is the version of the service
	Version = "v0.0.1"
	// GitSHA is the git sha this was built off
	GitSHA = "unknown"
)

// Config is the configuration for the service
type Config struct {
	// EnableClientTLS indicates you want mutual tls
	EnableClientTLS bool `yaml:"enable-client-tls"`
	// EnableLogging indicates you want http logging
	EnableLogging bool `yaml:"enable-logging"`
	// Listen is the interface we are listening on
	Listen string `yaml:"listen"`
	// TLSCert is the path to a certificate
	TLSCert string `yaml:"tls-cert"`
	// TLSKey is the path to a private key
	TLSKey string `yaml:"tls-key"`
	// TLSCA is the path to a ca
	TLSCA string `yaml:"tls-ca"`
	// Verbose indicates verbose logging
	Verbose bool `yaml:"verbose"`
}

// AdmissionReview describes an admission request.
type AdmissionReview struct {
	metav1.TypeMeta

	// Spec describes the attributes for the admission request.
	// Since this admission controller is non-mutating the webhook should avoid setting this in its response to avoid the
	// cost of deserializing it.
	Spec AdmissionReviewSpec
	// Status is filled in by the webhook and indicates whether the admission request should be permitted.
	Status AdmissionReviewStatus
}

// AdmissionReviewSpec describes the admission.Attributes for the admission request.
type AdmissionReviewSpec struct {
	// Kind is the type of object being manipulated.  For example: Pod
	Kind metav1.GroupVersionKind
	// Name is the name of the object as presented in the request.  On a CREATE operation, the client may omit name and
	// rely on the server to generate the name.  If that is the case, this method will return the empty string.
	Name string
	// Namespace is the namespace associated with the request (if any).
	Namespace string
	// Object is the object from the incoming request prior to default values being applied
	Object *extensions.Ingress
	// OldObject is the existing object. Only populated for UPDATE requests.
	OldObject *extensions.Ingress
	// Operation is the operation being performed
	Operation admission.Operation
	// Resource is the name of the resource being requested.  This is not the kind.  For example: pods
	Resource metav1.GroupVersionResource
	// SubResource is the name of the subresource being requested.  This is a different resource, scoped to the parent
	// resource, but it may have a different kind. For instance, /pods has the resource "pods" and the kind "Pod", while
	// /pods/foo/status has the resource "pods", the sub resource "status", and the kind "Pod" (because status operates on
	// pods). The binding resource for a pod though may be /pods/foo/binding, which has resource "pods", subresource
	// "binding", and kind "Binding".
	SubResource string
	// UserInfo is information about the requesting user
	UserInfo authentication.UserInfo
}

// AdmissionReviewStatus describes the status of the admission request.
type AdmissionReviewStatus struct {
	// Allowed indicates whether or not the admission request was permitted.
	Allowed bool
	// Result contains extra details into why an admission request was denied.
	// This field IS NOT consulted in any way if "Allowed" is "true".
	// +optional
	Result *metav1.Status
}
