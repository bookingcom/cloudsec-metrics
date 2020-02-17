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
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	g "github.com/jtaczanowski/go-graphite-client"
	"github.com/pkg/errors"

	"github.com/bookingcom/cloudsec-metrics/api"
	"github.com/bookingcom/cloudsec-metrics/graphite"
)

type opts struct {
	CollectPeriod          time.Duration `long:"collect_period" env:"COLLECT_PERIOD" default:"1m" description:"Time between metrics collection"`
	PrismAPIUrl            string        `long:"prisma_api_url" env:"PRISMA_API_URL" default:"https://api.eu.prismacloud.io" description:"Prisma API URL"`
	PrismAPIKey            string        `long:"prisma_api_key" env:"PRISMA_API_KEY" description:"Prisma API key"`
	PrismAPIPassword       string        `long:"prisma_api_password" env:"PRISMA_API_PASSWORD" description:"Prisma API password"`
	GraphiteHost           string        `long:"graphite_host" env:"GRAPHITE_HOST" description:"Graphite hostname"`
	GraphitePort           int           `long:"graphite_port" env:"GRAPHITE_PORT" default:"2003" description:"Graphite port"`
	GraphitePrefix         string        `long:"graphite_prefix" env:"GRAPHITE_PREFIX" description:"Graphite global prefix"`
	CompliancePrefix       string        `long:"compliance_prefix" env:"COMPLIANCE_PREFIX" default:"compliance." description:"Graphite compliance metrics prefix"`
	SCCDelayPrefix         string        `long:"scc_delay_prefix" env:"SCC_DELAY_PREFIX" default:"scc_delay." description:"Graphite SCC sources delay metrics prefix"`
	SCCHealthMetricName    string        `long:"scc_health_metric_name" env:"SCC_HEALTH_METRIC_NAME" default:"scc_health" description:"Graphite SCC health metric name"`
	PrismaHealthMetricName string        `long:"prisma_health_metric_name" env:"PRISMA_HEALTH_METRIC_NAME" default:"prisma_health" description:"Graphite Prisma health metric name"`
	SCCOrgID               string        `long:"scc_org_id" env:"SCC_ORG_ID" description:"Google SCC numeric organisation ID"`
	SCCSourcesRegex        string        `long:"scc_sources_regex" env:"SCC_SOURCES_REGEX" default:"." description:"Google SCC sources Display Name regexp"`
	Dbg                    bool          `long:"dbg" env:"DEBUG" description:"debug mode"`
}

type collectors struct {
	prisma     *api.Prisma
	sccSources map[string]string
}

type senders struct {
	graphite *g.Client
}

type metrics struct {
	complianceInfo        []api.ComplianceInfo
	googleSourcesDelay    map[string]time.Duration
	prismaHealthStatus    int
	googleSCCHealthStatus int
}

func main() {
	var opts = opts{}
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	log.SetFlags(log.Ldate | log.Ltime)
	if opts.Dbg {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	}

	var metrics = &metrics{}
	collectors, err := prepareCollectors(opts)
	if err != nil {
		log.Fatalf("[ERROR] Can't initialise collectors, %v", err)
	}
	senders := prepareSenders(opts)

	for ticker := time.NewTicker(opts.CollectPeriod); true; <-ticker.C {
		collectMetrics(metrics, collectors, "https://status.cloud.google.com/incidents.json")
		sendMetrics(metrics, senders, opts)
	}
}

// create and return a link to collectors with credentials provided via opts,
// return error in case of problems with connection initialisation
func prepareCollectors(opts opts) (*collectors, error) {
	var collectors = &collectors{}
	if opts.PrismAPIKey != "" && opts.PrismAPIPassword != "" {
		log.Printf("[INFO] Initialising Prisma data collection with API key %s", opts.PrismAPIKey)
		collectors.prisma = api.NewPrisma(opts.PrismAPIKey, opts.PrismAPIPassword, opts.PrismAPIUrl)
	}
	if opts.SCCOrgID != "" {
		var err error
		log.Printf("[INFO] Initialising Google Security Command Center data collection for Organisation ID %s", opts.SCCOrgID)
		collectors.sccSources, err = api.GetSCCSourcesByName(opts.SCCOrgID, opts.SCCSourcesRegex)
		if err != nil {
			return nil, errors.Wrap(err, "can't get SCC sources information")
		}
	}
	return collectors, nil
}

// create and return a pointer to senders
func prepareSenders(opts opts) *senders {
	var senders = &senders{}
	if opts.GraphiteHost != "" {
		senders.graphite = g.NewClient(opts.GraphiteHost, opts.GraphitePort, opts.GraphitePrefix, "tcp")
	}
	return senders
}

// collectMetrics collects metrics into referenced metrics object using provided collectors
func collectMetrics(metrics *metrics, collectors *collectors, googleHealthDashboard string) {
	var err error
	if collectors.prisma != nil {
		if metrics.complianceInfo, err = collectors.prisma.GatherComplianceInfo(); err != nil {
			log.Printf("[ERROR] Can't request compliance information, %v", err)
		}
		metrics.prismaHealthStatus = collectors.prisma.GetAPIHealthStatus()
	}
	if googleHealthDashboard != "" {
		metrics.googleSCCHealthStatus = api.GetSCCHealthStatus(googleHealthDashboard)
	}
	if collectors.sccSources != nil {
		if metrics.googleSourcesDelay, err = api.GetSCCLatestEventTime(collectors.sccSources); err != nil {
			log.Printf("[ERROR] Can't get SCC sources last update information, %v", err)
		}
	}
}

// sendMetrics sends metrics to initialised senders
func sendMetrics(metrics *metrics, senders *senders, opts opts) {
	if senders.graphite != nil {
		graphiteMetrics := map[string]float64{}
		if metrics.complianceInfo != nil {
			for k, v := range graphite.GenerateComplianceInfo(opts.CompliancePrefix, metrics.complianceInfo) {
				graphiteMetrics[k] = v
			}
			graphiteMetrics[opts.PrismaHealthMetricName] = float64(metrics.prismaHealthStatus)
		}
		for k, v := range graphite.GenerateSSCSourcesDelay(opts.SCCDelayPrefix, metrics.googleSourcesDelay) {
			graphiteMetrics[k] = v
		}
		graphiteMetrics[opts.SCCHealthMetricName] = float64(metrics.googleSCCHealthStatus)
		if err := senders.graphite.SendData(graphiteMetrics); err != nil {
			log.Printf("[ERROR] Can't send metrics to Graphite, %v", err)
		}
	}
}
