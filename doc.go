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

const (
	// AdmissionControllerName is the name we register as
	AdmissionControllerName = "acp-io.ingress.control.homeoffice.gov.uk"
	// DomainWhitelistAnnotation is the annotation which controls which domains you can use
	DomainWhitelistAnnotation = "acp.io/homeoffice/ingress-domains"
)

var (
	// Version is the version of the service
	Version = "v0.0.1"
	// GitSHA is the git sha this was built off
	GitSHA = "unknown"
)

// Config is the configuration for the service
type Config struct {
	// Listen is the interface we are listening on
	Listen string `yaml:"listen"`
	// TLSCert is the path to a certificate
	TLSCert string `yaml:"tls-cert"`
	// TLSKey is the path to a private key
	TLSKey string `yaml:"tls-key"`
	// TLSCA is the path to a ca
	TLSCA string `yaml:"tls-ca"`
	// EnableClientTLS indicates you want mutual tls
	EnableClientTLS bool `yaml:"enable-client-tls"`
	// Verbose indicates verbose logging
	Verbose bool `yaml:"verbose"`
}
