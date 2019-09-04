// Copyright 2019 Booking.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSCCHealthStatus(t *testing.T) {
	endedEventServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/incidents.json")
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, `
[{"service_key":"cloud-security-command-center",
"description":"ended incident",
"begin":"1999-01-01T00:00:00Z",
"end":"1999-01-01T00:01:00Z"
}]`)
	}))
	defer endedEventServer.Close()
	ongoingEventServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, `
		[{"service_key": "cloud-security-command-center",
			"description": "ongoing incident",
			"begin":       "1999-01-01T00:00:00Z",
			"end":         "2999-01-01T00:01:00Z"
		}]`)
	}))
	defer ongoingEventServer.Close()
	badResponseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Length", "1")
	}))
	defer badResponseServer.Close()
	serverStatusUnauthorized := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer serverStatusUnauthorized.Close()
	serverStatusNotFound := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer serverStatusNotFound.Close()

	var testAPIRequestsDataset = []struct {
		serverURL string
		status    int
	}{
		{serverURL: "http://[::1]:namedport"},
		{serverURL: "nonexistent_url"},
		{serverURL: endedEventServer.URL + "/incidents.json", status: 1},
		{serverURL: badResponseServer.URL},
		{serverURL: serverStatusNotFound.URL},
		{serverURL: ongoingEventServer.URL},
	}

	for i, x := range testAPIRequestsDataset {
		status := GetSCCHealthStatus(x.serverURL)
		assert.Equal(t, x.status, status, "Test case %d status check failed", i)
	}
}
