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
	"time"

	"github.com/bookingcom/cloudsec-metrics/api"
)

// GenerateComplianceInfo returns metrics from given compliance info
func GenerateComplianceInfo(prefix string, ci []api.ComplianceInfo) map[string]float64 {
	metrics := map[string]float64{}
	for _, entry := range ci {
		metricPrefix := prefix + escapeMetricName(entry.Name)
		metrics[metricPrefix+".policies_total"] = float64(entry.PoliciesCount)
		metrics[metricPrefix+".assets_passed"] = float64(entry.PassedAssetsCount)
		metrics[metricPrefix+".assets_failed"] = float64(entry.FailedAssetsCount)
		metrics[metricPrefix+".assets_total"] = float64(entry.TotalAssetsCount)
	}
	return metrics
}

// GenerateSSCSourcesDelay returns metrics from given delay map
func GenerateSSCSourcesDelay(prefix string, delay map[string]time.Duration) map[string]float64 {
	metrics := map[string]float64{}
	for name, duration := range delay {
		metricPrefix := prefix + escapeMetricName(name)
		metrics[metricPrefix+".seconds"] = duration.Seconds()
	}
	return metrics
}

func escapeMetricName(name string) string {
	result := ""
	for _, c := range name {
		switch c {
		case ' ', '.', '{', '}', '(', ')', '/':
			result += "_"
		default:
			result += string(c)
		}
	}
	return result
}
