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

package graphite

import (
	"testing"

	"github.com/bookingcom/cloudsec-metrics/api"
	"github.com/stretchr/testify/assert"
)

func TestGraphite(t *testing.T) {
	var testDataset = []struct {
		host                   string
		port                   int
		prefix, expectedPrefix string
	}{
		{},
		{host: "bad_host", prefix: "some_prefix"},
	}

	for i, x := range testDataset {
		g := New(x.host, x.port, x.prefix)
		assert.Equal(t, x.host, g.Host, "Test case %d host check failed", i)
		assert.Equal(t, x.port, g.Port, "Test case %d port check failed", i)
		assert.Equal(t, x.expectedPrefix, g.Prefix, "Test case %d prefix check failed", i)
	}

	g := New("", 0, "")
	assert.NoError(t, SendComplianceInfo(g, "", nil), "Empty metric send should do nothing and return no errors")
	assert.NoError(t, SendComplianceInfo(g, "", []api.ComplianceInfo{{Name: "test_name"}}), "Empty metric send should do nothing and return no errors")
}
