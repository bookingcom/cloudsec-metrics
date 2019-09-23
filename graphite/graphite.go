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
	"github.com/pkg/errors"
)

// New is creating new Graphite instance with given parameters, disabling graphite.Graphite default logging
func New(host string, port int, prefix string) (G *graphite.Graphite, err error) {
	if G, err = graphite.GraphiteFactory("tcp", host, port, prefix); err != nil {
		return nil, errors.Wrap(err, "can't connect to Graphite")
	}
	G.DisableLog = true
	return G, nil
}

// SendComplianceInfo tries to send assets compliance information to graphite
func SendComplianceInfo(g *graphite.Graphite, prefix string, ci []api.ComplianceInfo) error {
	timeNow := time.Now().Unix()
	var metrics []graphite.Metric
	for _, x := range ci {
		metricPrefix := escapeMetricName(prefix + x.Name)
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".policies_total", strconv.Itoa(x.PoliciesCount), timeNow))
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".assets_passed", strconv.Itoa(x.PassedAssetsCount), timeNow))
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".assets_failed", strconv.Itoa(x.FailedAssetsCount), timeNow))
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".assets_total", strconv.Itoa(x.TotalAssetsCount), timeNow))
	}
	return g.SendMetrics(metrics)
}

// SendSSCSourcesDelay tries to send SCC sources delay information to graphite
func SendSSCSourcesDelay(g *graphite.Graphite, prefix string, delay map[string]time.Duration) error {
	timeNow := time.Now().Unix()
	var metrics []graphite.Metric
	for k, v := range delay {
		metricPrefix := escapeMetricName(prefix + k)
		metrics = append(metrics, graphite.NewMetric(metricPrefix+".policies_total", fmt.Sprintf("%f", v.Seconds()), timeNow))
	}
	return g.SendMetrics(metrics)
}

// SendMetric tries to send given metric to graphite
func SendMetric(g *graphite.Graphite, metric string, value string) error {
	timeNow := time.Now().Unix()
	metrics := []graphite.Metric{graphite.NewMetric(escapeMetricName(metric), value, timeNow)}
	return g.SendMetrics(metrics)
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
