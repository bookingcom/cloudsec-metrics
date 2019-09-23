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
	"time"

	"github.com/bookingcom/cloudsec-metrics/api"
	"github.com/marpaia/graphite-golang"
	"github.com/stretchr/testify/assert"
)

func TestGraphite(t *testing.T) {
	var testDataset = []struct {
		host                   string
		port                   int
		prefix, expectedPrefix string
		nilReturn              bool
	}{
		{nilReturn: true},
		{host: "bad_host"},
	}

	for i, x := range testDataset {
		g, err := New(x.host, x.port, x.prefix)
		assert.Error(t, err, "Test case %d error check failed", i)
		assert.Nil(t, g, "Test case %d nil object check failed", i)
	}

	assert.NoError(t, SendComplianceInfo(&graphite.Graphite{}, "", nil),
		"Empty metric send should do nothing and return no errors")
	assert.NoError(t, SendComplianceInfo(&graphite.Graphite{}, "", []api.ComplianceInfo{{Name: "test{name}"}}),
		"Single metric send to empty Graphite should do nothing and return no errors")
	assert.NoError(t, SendSSCSourcesDelay(&graphite.Graphite{}, "", nil),
		"Empty metric send should do nothing and return no errors")
	assert.NoError(t, SendSSCSourcesDelay(&graphite.Graphite{}, "", map[string]time.Duration{"test": time.Second}),
		"Single metric send to empty Graphite should do nothing and return no errors")
	assert.NoError(t, SendMetric(&graphite.Graphite{}, "", ""),
		"Single metric send should do nothing and return no errors")
	assert.Equal(t, "_test_of_metric", escapeMetricName("(test)of/metric"))
}
