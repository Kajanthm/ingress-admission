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
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

type request struct {
	Method          string
	URI             string
	Body            string
	ExpectedCode    int
	ExpectedContent string
}

type fakeController struct {
	svc *httptest.Server
	ctl *controller
}

func newFakeController() *fakeController {
	log.SetOutput(ioutil.Discard)
	c, _ := newController(Config{
		EnableLogging: false,
	})
	c.client = fake.NewSimpleClientset()

	return &fakeController{svc: httptest.NewServer(c.engine), ctl: c}
}

// runTests performs a series of tests on the service
func (c *fakeController) runTests(t *testing.T, requests []request) {
	for i, x := range requests {
		// set the sane defaults
		method := http.MethodGet
		if x.Method != "" {
			method = x.Method
		}
		req, err := http.NewRequest(method, c.svc.URL+x.URI, bytes.NewBufferString(x.Body))
		require.NoError(t, err, "case %d should not have thrown error: %s", i, err)
		require.NotNil(t, req, "case %d response should not be nil", i)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "case %d should not have thrown error: %s", i, err)
		require.NotNil(t, resp, "case %d response should not be nil", i)
		content, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "case %d unable to read content, error: %s", i, err)

		if x.ExpectedCode != 0 {
			assert.Equal(t, x.ExpectedCode, resp.StatusCode, "case %d, expected: %d, got: %d", i, x.ExpectedCode, resp.StatusCode)
		}
		if x.ExpectedContent != "" {
			assert.Equal(t, x.ExpectedContent, string(content), "case %d, expected: %s, got: %s", i, x.ExpectedContent, string(content))
		}
	}
}
