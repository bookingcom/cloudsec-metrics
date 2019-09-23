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
	"fmt"
	"strconv"
	"time"

	"github.com/bookingcom/cloudsec-metrics/api"
	"github.com/marpaia/graphite-golang"
)

// GenerateComplianceInfo returns metrics from given compliance info
func GenerateComplianceInfo(timeNow int64, prefix string, ci []api.ComplianceInfo) []graphite.Metric {
	var metrics []graphite.Metric
	for _, x := range ci {
		metricPrefix := prefix + escapeMetricName(x.Name)
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".policies_total", strconv.Itoa(x.PoliciesCount), timeNow))
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".assets_passed", strconv.Itoa(x.PassedAssetsCount), timeNow))
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".assets_failed", strconv.Itoa(x.FailedAssetsCount), timeNow))
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".assets_total", strconv.Itoa(x.TotalAssetsCount), timeNow))
	}
	return metrics
}

// GenerateSSCSourcesDelay returns metrics from given delay map
func GenerateSSCSourcesDelay(timeNow int64, prefix string, delay map[string]time.Duration) []graphite.Metric {
	var metrics []graphite.Metric
	for k, v := range delay {
		metricPrefix := prefix + escapeMetricName(k)
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".policies_total", fmt.Sprintf("%f", v.Seconds()), timeNow))
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
