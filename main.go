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

	"github.com/bookingcom/cloudsec-metrics/api"
	"github.com/bookingcom/cloudsec-metrics/graphite"
	"github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		PrismAPIUrl      string `long:"prisma_api_url" env:"PRISMA_API_URL" default:"https://api.eu.prismacloud.io" description:"Prisma API URL"`
		PrismAPIKey      string `long:"prisma_api_key" env:"PRISMA_API_KEY" description:"Prisma API key"`
		PrismAPIPassword string `long:"prisma_api_password" env:"PRISMA_API_PASSWORD" description:"Prisma API password"`
		GraphiteHost     string `long:"graphite_host" env:"GRAPHITE_HOST" description:"Graphite hostname"`
		GraphitePort     int    `long:"graphite_port" env:"GRAPHITE_PORT" default:"2003" description:"Graphite port"`
		GraphitePrefix   string `long:"graphite_prefix" env:"GRAPHITE_PREFIX" description:"Graphite global prefix"`
		CompliancePrefix string `long:"compliance_prefix" env:"COMPLIANCE_PREFIX" default:"compliance." description:"Graphite compliance metrics prefix"`
		SCCOrgID         string `long:"scc_org_id" env:"SCC_ORG_ID" description:"Google SCC numeric organisation ID"`
		SCCSourcesRegex  string `long:"scc_sources_regex" env:"SCC_SOURCES_REGEX" default:"." description:"Google SCC sources Display Name regexp"`
		Dbg              bool   `long:"dbg" env:"DEBUG" description:"debug mode"`
	}

	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	log.SetFlags(log.Ldate | log.Ltime)
	if opts.Dbg {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	}

	var collectedMetrics struct {
		complianceInfo        []api.ComplianceInfo
		googleSourcesDelay    map[string]time.Duration
		prismaHealthStatus    int
		googleSCCHealthStatus int
	}

	if opts.PrismAPIKey != "" && opts.PrismAPIPassword != "" {
		log.Print("[INFO] Starting Prisma data collection")
		prisma, err := api.NewPrisma(opts.PrismAPIKey, opts.PrismAPIPassword, opts.PrismAPIUrl)
		if err != nil {
			log.Fatalf("[ERROR] Can't connect to Prisma, %v", err)
		}
		collectedMetrics.complianceInfo, err = prisma.GatherComplianceInfo()
		if err != nil {
			log.Printf("[ERROR] Can't request complience information, %v", err)
		}
		collectedMetrics.prismaHealthStatus = prisma.GetAPIHealthStatus()
	}
	collectedMetrics.googleSCCHealthStatus = api.GetSCCHealthStatus("https://status.cloud.google.com/incidents.json")

	Graphite := graphite.New(opts.GraphiteHost, opts.GraphitePort, opts.GraphitePrefix)
	if err := graphite.SendComplianceInfo(Graphite, opts.CompliancePrefix, collectedMetrics.complianceInfo); err != nil {
		log.Printf("[ERROR] Can't send complience information, %v", err)
	}
	if opts.SCCOrgID != "" {
		sccSources, err := api.GetSCCSourcesByName(opts.SCCOrgID, opts.SCCSourcesRegex)
		if err != nil {
			log.Fatalf("[ERROR] Can't get SCC sources information, %v", err)
		}
		if collectedMetrics.googleSourcesDelay, err = api.GetSCCLatestEventTime(sccSources); err != nil {
			log.Printf("[ERROR] Can't get SCC sources last update information, %v", err)
		}
	}
}
