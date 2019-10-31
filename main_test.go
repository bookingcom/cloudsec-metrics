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

	"github.com/jtaczanowski/go-graphite-client"
	"github.com/stretchr/testify/assert"

	"github.com/bookingcom/cloudsec-metrics/api"
)

func TestPrepareCollectors(t *testing.T) {
	var testDataset = []struct {
		opts       opts
		err        bool
		collectors *collectors
	}{
		{collectors: &collectors{}},
		{collectors: &collectors{
			prisma: &api.Prisma{Login: "bad", Password: "bad_pass", APIUrl: "bad_host"}},
			opts: opts{PrismAPIKey: "bad", PrismAPIPassword: "bad_pass", PrismAPIUrl: "bad_host"}},
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
	s := prepareSenders(opts{})
	assert.Equal(t, &senders{}, s, "No senders initialised without options provided")
}

func TestCollectMetrics(t *testing.T) {
	m := metrics{}
	collectMetrics(&m, &collectors{}, "")
	assert.Equal(t, metrics{}, m, "No data collected without collectors provided")
}

func TestSendMetrics(t *testing.T) {
	m := metrics{complianceInfo: []api.ComplianceInfo{}}
	sendMetrics(&m, &senders{graphite: &graphite.Client{}}, opts{})
	assert.Equal(t, metrics{complianceInfo: []api.ComplianceInfo{}}, m, "Metrics unchanged after send function call")
}
