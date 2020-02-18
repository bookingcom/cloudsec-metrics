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
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPrisma_GatherComplianceInfo(t *testing.T) {
	var testAPIRequestsDataset = []struct {
		serverErr error
		error     string
		answer    []byte
		asset     []ComplianceInfo
	}{
		{serverErr: errors.New("mock error"),
			error: "error requesting assets information: mock error"},
		{answer: []byte(`{"timestamp": 1571919534777,"complianceDetails":[{"name":"test_name","description":"test description","passedResources":69,"assignedPolicies":66,"failedResources":99, "totalResources":168}]}`),
			asset: []ComplianceInfo{{"test_name", "test description", 66, 69, 99, 69 + 99}}},
		{answer: []byte("not_json"),
			error: "error unmarshaling assets information: invalid character 'o' in literal null (expecting 'u')"},
	}

	// start tests
	p := &Prisma{}

	for i, x := range testAPIRequestsDataset {
		p.api = &mockClient{t: t, url: "/compliance/posture?timeType=to_now&timeUnit=day", method: "GET",
			err: x.serverErr, answer: x.answer}
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
	var testAPIRequestsDataset = []struct {
		err    error
		status int
	}{
		{err: errors.New("mock problem")},
		{status: 1},
	}

	// start tests
	p := &Prisma{}

	for i, x := range testAPIRequestsDataset {
		p.api = &mockClient{t: t, url: "/check", method: "GET", err: x.err}
		status := p.GetAPIHealthStatus()
		assert.Equal(t, x.status, status, "Test case %d status code check failed", i)
	}
}

type mockClient struct {
	t      *testing.T
	method string
	url    string
	answer []byte
	err    error
}

func (m *mockClient) DoAPIRequest(method, url string, _ io.Reader) ([]byte, error) {
	assert.Equal(m.t, m.url, url)
	assert.Equal(m.t, m.method, method)
	return m.answer, m.err
}
