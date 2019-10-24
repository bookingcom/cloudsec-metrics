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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiRequest(t *testing.T) {
	// prepare servers
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/login", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		assert.Equal(t, "test_text", buf.String())
		_, _ = fmt.Fprint(w, "one, two, three")
	}))
	defer goodServer.Close()
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Length", "1")
	}))
	defer badServer.Close()
	serverStatusUnauthorized := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer serverStatusUnauthorized.Close()
	serverStatusBadRequest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer serverStatusBadRequest.Close()
	serverStatusInternalServerError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer serverStatusInternalServerError.Close()
	serverStatusNotFound := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer serverStatusNotFound.Close()

	var testAPIRequestsDataset = []struct {
		serverURL    string
		method       string
		url          string
		error        string
		responseBody []byte
		body         io.Reader
	}{
		{serverURL: "http://[::1]:namedport", method: "POST",
			error: "error creating request: parse http://[::1]:namedport: invalid port \":namedport\" after host"},
		{serverURL: "nonexistent_url", method: "POST",
			error: "error getting auth token: error logging in with user \"\": error making request: Post nonexistent_url/login: unsupported protocol scheme \"\""},
		{serverURL: goodServer.URL, method: "POST", url: "/login",
			responseBody: []byte("one, two, three"), body: bytes.NewReader([]byte("test_text"))},
		{serverURL: badServer.URL, method: "GET", url: "/",
			error: "error reading response body, response body: \"\": unexpected EOF"},
		{serverURL: serverStatusUnauthorized.URL, method: "GET",
			error: "authentication error on request, response body: \"\""},
		{serverURL: serverStatusBadRequest.URL, method: "GET",
			error: "bad request parameters, check your request body, response body: \"\""},
		{serverURL: serverStatusInternalServerError.URL, method: "GET",
			error: "server internal error during request processing, response body: \"\""},
		{serverURL: serverStatusNotFound.URL, method: "GET",
			error: "404 Not Found, response body: \"\""},
	}

	// start tests
	p := &Prisma{}

	for i, x := range testAPIRequestsDataset {
		p.APIUrl = x.serverURL
		data, err := p.doAPIRequest(x.method, x.url, x.body)
		if x.error != "" {
			assert.EqualError(t, err, x.error, "Test case %d error check failed", i)
		} else {
			assert.NoError(t, err, "Test case %d error check failed", i)
		}
		if x.responseBody != nil {
			assert.Equal(t, x.responseBody, data, "Test case %d response data check failed", i)
		} else {
			assert.Nil(t, data, "Test case %d response data check failed", i)
		}
	}
}

func TestNewPrisma(t *testing.T) {
	// prepare servers
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/login", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		assert.Equal(t, "{\"password\":\"test_password\",\"username\":\"test_user\"}", buf.String())
		_, _ = fmt.Fprint(w, "{\"token\":\"test_token\"}")
	}))
	defer goodServer.Close()
	goodRenewServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/auth_token/extend", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "{\"token\":\"test_token_renewed\"}")
	}))
	defer goodRenewServer.Close()
	badRenewServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer badRenewServer.Close()
	badEmptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer badEmptyServer.Close()
	badAnswerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/login")
		assert.Equal(t, "POST", r.Method)
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		assert.Equal(t, "{\"password\":\"\",\"username\":\"\"}", buf.String())
		_, _ = fmt.Fprint(w, "not_json")
	}))
	defer badAnswerServer.Close()

	var testAPIRequestsDataset = []struct {
		serverURL     string
		username      string
		password      string
		error         string
		responseToken string
		setToken      string
	}{
		{serverURL: "nonexistent_url", username: "test_username",
			error: "error logging in with user \"test_username\": error making request: Post nonexistent_url/login: unsupported protocol scheme \"\""},
		{serverURL: goodServer.URL, username: "test_user", password: "test_password", responseToken: "test_token"},
		{serverURL: badAnswerServer.URL,
			error: "error obtaining token from login response: invalid character 'o' in literal null (expecting 'u')"},
		{serverURL: goodRenewServer.URL, username: "test_user", password: "test_password",
			setToken: "old_good_token", responseToken: "test_token_renewed"},
		{serverURL: badRenewServer.URL, username: "test_user", password: "test_password", setToken: "old_bad_token",
			error: "error logging in with user \"test_user\": authentication error on request, response body: \"\""},
		{serverURL: badEmptyServer.URL, username: "test_user", password: "test_password", setToken: "old_bad_token",
			error: "error obtaining token from login response: unexpected end of JSON input"},
	}

	// start tests

	for i, x := range testAPIRequestsDataset {
		p := &Prisma{Login: x.username, Password: x.password, APIUrl: x.serverURL}
		p.Token = x.setToken
		err := p.authenticate()
		if x.error != "" {
			assert.EqualError(t, err, x.error, "Test case %d error check failed", i)
		} else {
			assert.NoError(t, err, "Test case %d error check failed", i)
		}
		assert.NotNil(t, p, "Test case %d Prisma object return check failed", i)
		assert.Equal(t, x.responseToken, p.Token, "Test case %d Prisma object token check failed", i)
	}
}

func TestPrisma_GatherComplianceInfo(t *testing.T) {
	// prepare servers
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/compliance/posture", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "{\"timestamp\": 1571919534777,\"complianceDetails\":[{\"name\":\"test_name\",\"description\":\"test description\",\"passedResources\":69,\"assignedPolicies\":66,\"failedResources\":99, \"totalResources\":168}]}")
	}))
	defer goodServer.Close()
	badAnswerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/compliance/posture", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, "not_json")
	}))
	defer badAnswerServer.Close()

	var testAPIRequestsDataset = []struct {
		serverURL string
		error     string
		asset     []ComplianceInfo
	}{
		{serverURL: "nonexistent_url",
			error: "error requesting assets information: error getting auth token: error logging in with user \"\": error making request: Post nonexistent_url/login: unsupported protocol scheme \"\""},
		{serverURL: goodServer.URL,
			asset: []ComplianceInfo{{"test_name", "test description", 66, 69, 99, 69 + 99}}},
		{serverURL: badAnswerServer.URL,
			error: "error unmarshaling assets information: invalid character 'o' in literal null (expecting 'u')"},
	}

	// start tests
	p := &Prisma{}

	for i, x := range testAPIRequestsDataset {
		p.APIUrl = x.serverURL
		assetInfo, err := p.GatherComplianceInfo()
		if x.error != "" {
			assert.EqualError(t, err, x.error, "Test case %d error check failed", i)
		} else {
			assert.NoError(t, err, "Test case %d error check failed", i)
		}
		assert.Equal(t, x.asset, assetInfo, "Test case %d assetInfo object check failed", i)
	}
}

func TestPrisma_GetAPIHealthStatus(t *testing.T) {
	// prepare servers
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/check", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
	}))
	defer goodServer.Close()
	badAnswerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badAnswerServer.Close()

	var testAPIRequestsDataset = []struct {
		serverURL string
		status    int
	}{
		{serverURL: "nonexistent_url"},
		{serverURL: goodServer.URL, status: 1},
		{serverURL: badAnswerServer.URL},
	}

	// start tests
	p := &Prisma{}

	for i, x := range testAPIRequestsDataset {
		p.APIUrl = x.serverURL
		status := p.GetAPIHealthStatus()
		assert.Equal(t, x.status, status, "Test case %d status code check failed", i)
	}
}
