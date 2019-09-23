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

package main

import (
	"testing"

	"github.com/bookingcom/cloudsec-metrics/api"
	"github.com/marpaia/graphite-golang"
	"github.com/stretchr/testify/assert"
)

func TestPrepareCollectors(t *testing.T) {
	var testDataset = []struct {
		opts       opts
		err        bool
		collectors *collectors
	}{
		{collectors: &collectors{}},
		{opts: opts{PrismAPIKey: "bad", PrismAPIPassword: "bad", PrismAPIUrl: "bad_host"}, err: true},
		{opts: opts{SCCOrgID: "bad"}, err: true},
	}
	for i, x := range testDataset {
		c, err := prepareCollectors(x.opts)
		if x.err {
			assert.Error(t, err, "Test case %d error check failed", i)
		} else {
			assert.NoError(t, err, "Test case %d error check failed", i)
		}
		assert.Equal(t, x.collectors, c, "Test case %d collectors check failed", i)
	}
}

func TestPrepareSenders(t *testing.T) {
	s, err := prepareSenders(opts{})
	assert.NoError(t, err)
	assert.Equal(t, &senders{}, s, "No senders initialised without options provided")
	s, err = prepareSenders(opts{GraphiteHost: "bad_host"})
	assert.Error(t, err, "Bad Graphite hostname results in error")
	assert.Nil(t, s, "Bad Graphite hostname results in no senders created")
}

func TestCollectMetrics(t *testing.T) {
	m := metrics{}
	collectMetrics(&m, &collectors{}, "")
	assert.Equal(t, metrics{}, m, "No data collected without collectors provided")
}

func TestSendMetrics(t *testing.T) {
	m := metrics{complianceInfo: []api.ComplianceInfo{}}
	sendMetrics(&m, &senders{graphite: &graphite.Graphite{}}, opts{})
	assert.Equal(t, metrics{complianceInfo: []api.ComplianceInfo{}}, m, "Metrics unchanged after send function call")
}
