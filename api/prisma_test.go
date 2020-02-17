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
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPrisma_GatherComplianceInfo(t *testing.T) {
	// prepare servers
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/compliance/posture", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, `{"timestamp": 1571919534777,"complianceDetails":[{"name":"test_name","description":"test description","passedResources":69,"assignedPolicies":66,"failedResources":99, "totalResources":168}]}`)
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
			error: "error requesting assets information: error making request: Get nonexistent_url/compliance/posture?timeType=to_now&timeUnit=day: unsupported protocol scheme \"\""},
		{serverURL: goodServer.URL,
			asset: []ComplianceInfo{{"test_name", "test description", 66, 69, 99, 69 + 99}}},
		{serverURL: badAnswerServer.URL,
			error: "error unmarshaling assets information: invalid character 'o' in literal null (expecting 'u')"},
	}

	// start tests
	p := &Prisma{}

	for i, x := range testAPIRequestsDataset {
		p.api = &mockClient{APIURL: x.serverURL}
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
		p.api = &mockClient{APIURL: x.serverURL}
		status := p.GetAPIHealthStatus()
		assert.Equal(t, x.status, status, "Test case %d status code check failed", i)
	}
}

type mockClient struct{ APIURL string }

func (m *mockClient) DoAPIRequest(method, url string, body io.Reader) ([]byte, error) {
	req, _ := http.NewRequest(method, m.APIURL+url, body)
	req.Header.Set("Content-Type", "application/json")
	httpClient := http.Client{Timeout: time.Second}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error making request")
	}
	data, _ := ioutil.ReadAll(response.Body)
	_ = response.Body.Close()
	switch response.StatusCode {
	case http.StatusInternalServerError:
		return nil, errors.Errorf("server internal error during request processing, response body: %q", data)
	default:
		return data, nil
	}
}
