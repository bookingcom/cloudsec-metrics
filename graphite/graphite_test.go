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
	assert.Nil(t, GenerateComplianceInfo(0, "", nil),
		"Run with no metrics should return nil")
	assert.Equal(t,
		[]graphite.Metric{
			{Name: "test_name_.policies_total", Value: "1", Timestamp: 0},
			{Name: "test_name_.assets_passed", Value: "2", Timestamp: 0},
			{Name: "test_name_.assets_failed", Value: "3", Timestamp: 0},
			{Name: "test_name_.assets_total", Value: "4", Timestamp: 0}},
		GenerateComplianceInfo(0, "", []api.ComplianceInfo{
			{Name: "test{name}",
				PoliciesCount:     1,
				PassedAssetsCount: 2,
				FailedAssetsCount: 3,
				TotalAssetsCount:  4}}),
		"Single metric send to empty Graphite should do nothing and return no errors")
	assert.Nil(t, GenerateSSCSourcesDelay(0, "", nil),
		"Run with no metrics should return nil")
	assert.Equal(t,
		[]graphite.Metric{{Name: "test.policies_total", Value: "0.065000", Timestamp: 0}},
		GenerateSSCSourcesDelay(0, "", map[string]time.Duration{"test": time.Millisecond * 65}),
		"Single metric send to empty Graphite should do nothing and return no errors")
	assert.Equal(t, "_test_of_metric", escapeMetricName("(test)of/metric"))
}
