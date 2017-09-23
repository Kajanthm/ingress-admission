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
	"net/http"
	"testing"
)

type unitTest struct {
	Expected    string `yaml:"expected"`
	Request     string `yaml:"request"`
	Description string `yaml:"description"`
}

/*
func TestReviewHandler(t *testing.T) {
	var unitTests []unitTest

	content, err := ioutil.ReadFile("tests/unit-tests.yaml")
	require.NoError(t, err, "unable to read in the unit tests")
	require.NoError(t, yaml.Unmarshal(content, &unitTests))

	var requests []request

	c := newFakeController()
	c.ctl.client.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	})

	for _, x := range unitTests {
		requests = append(requests, request{
			URI:             "/",
			Method:          http.MethodPost,
			Body:            x.Request,
			ExpectedCode:    http.StatusOK,
			ExpectedContent: x.Expected,
		})
	}

	c.runTests(t, requests)
}
*/

func TestVersionHandler(t *testing.T) {
	requests := []request{
		{
			URI:             "/version",
			ExpectedCode:    http.StatusOK,
			ExpectedContent: Version,
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
