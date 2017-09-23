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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTLSCOnfig(t *testing.T) {
	c, err := getTLSConfig(&Config{})
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestHasDomainOK(t *testing.T) {
	cs := []struct {
		Hostname  string
		Whitelist []string
	}{
		{
			Hostname:  "test.web.svc.cluster.local",
			Whitelist: []string{"test.web.svc.cluster.local"},
		},
		{
			Hostname:  "test.web.svc.cluster.local",
			Whitelist: []string{"service.web.svc.cluster.local", "test.web.svc.cluster.local"},
		},
		{
			Hostname:  "test.web.svc.cluster.local",
			Whitelist: []string{"*.web.svc.cluster.local", "host.web.svc.cluster.local"},
		},
	}
	for i, c := range cs {
		assert.True(t, hasDomain(c.Hostname, c.Whitelist), "case %d, should have been true", i)
	}
}

func TestHasDomainBad(t *testing.T) {
	cs := []struct {
		Hostname  string
		Whitelist []string
	}{
		{
			Hostname:  "one.test.web.svc.cluster.local",
			Whitelist: []string{"*.web.svc.cluster.local", "host.web.svc.cluster.local"},
		},
	}
	for i, c := range cs {
		assert.False(t, hasDomain(c.Hostname, c.Whitelist), "case %d, should have been false", i)
	}
}
