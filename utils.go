package main

import (
	"crypto/tls"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

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

// getTLSConfig builds the TLS configuration from the options
func getTLSConfig(c *Config) (*tls.Config, error) {
	cfg := &tls.Config{
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
		ClientAuth:               tls.NoClientCert,
	}

	// @step: load the server certificates
	if c.TLSCert != "" && c.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(c.TLSCert, c.TLSKey)
		if err != nil {
			return nil, err
		}
		cfg.Certificates = []tls.Certificate{cert}
	}

	return cfg, nil
}

// getKubernetesClient returns a kubernetes api client for us
func getKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
