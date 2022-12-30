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

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	securitycenter "cloud.google.com/go/securitycenter/apiv1"
	"cloud.google.com/go/securitycenter/apiv1/securitycenterpb"
	"google.golang.org/api/iterator"
)

type googleStatusEntry struct {
	Service     string    `json:"service_key"`
	Description string    `json:"external_desc"`
	StartDate   time.Time `json:"begin"`
	EndDate     time.Time `json:"end"`
}

// Google SCC API operations timeout
const apiTimeout = time.Second * 20

// GetSCCHealthStatus gets Google Security Command Center health information and returns 1 on healthy response, 0 otherwise
// Check is performed by fetching list of incidents from Google Cloud Status Dashboard
// and checking if there are ongoing incidents with cloud-security-command-center;
// url parameter should be set to https://status.cloud.google.com/incidents.json for proper result retrieval
func GetSCCHealthStatus(url string) int {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0
	}
	httpClient := http.Client{Timeout: time.Second * 5}
	response, err := httpClient.Do(req)
	if err != nil {
		return 0
	}
	data, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return 0
	}
	var results []googleStatusEntry
	if err := json.Unmarshal(data, &results); err != nil {
		return 0
	}
	for _, entry := range results {
		if entry.Service == "cloud-security-command-center" &&
			entry.StartDate.Before(time.Now()) &&
			(entry.EndDate.Equal(time.Time{}) || entry.EndDate.After(time.Now())) {
			log.Printf("[INFO] Google Security Command Center incident in process since %v: %q",
				entry.StartDate, entry.Description)
			return 0
		}
	}
	return 1
}

// GetSCCSourcesByName returns Security Command Center sources for given numeric orgID,
// filtered by name by given regex
// original: https://github.com/GoogleCloudPlatform/golang-samples/blob/master/securitycenter/findings/list_sources.go
func GetSCCSourcesByName(orgID string, nameRegex string) (map[string]string, error) {
	regex, err := regexp.Compile(nameRegex)
	if err != nil {
		return nil, fmt.Errorf("error compiling nameRegex: %w", err)
	}
	// Instantiate a context and a security service client to make API calls.
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	client, err := securitycenter.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("securitycenter.NewClient: %w", err)
	}
	defer client.Close() // Closing the client safely cleans up background resources.

	req := &securitycenterpb.ListSourcesRequest{
		Parent: fmt.Sprintf("organizations/%s", orgID),
	}
	it := client.ListSources(ctx, req)
	result := map[string]string{}
	for {
		source, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("sources iterator problem: %w", err)
		}

		if match := regex.MatchString(source.DisplayName); match {
			result[source.Name] = source.DisplayName
		}
	}
	return result, nil
}

// GetSCCLatestEventTime return map of sources and their latest event update time difference with now
// original: https://github.com/GoogleCloudPlatform/golang-samples/blob/master/securitycenter/findings/list_filtered_findings.go
func GetSCCLatestEventTime(sources map[string]string) (map[string]time.Duration, error) {
	result := make(map[string]time.Duration)
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()
	client, err := securitycenter.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("securitycenter.NewClient: %w", err)
	}
	defer client.Close() // Closing the client safely cleans up background resources.

	// process just one event with newest update date for every given source
	for id, name := range sources {
		req := &securitycenterpb.ListFindingsRequest{
			Parent:   id,
			OrderBy:  `event_time desc`,
			PageSize: 1,
		}
		it := client.ListFindings(ctx, req)
		// we are getting first page with single element and discard other results
		findingsResult, err := it.Next()
		if err == iterator.Done {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("events iterator problem: %w", err)
		}
		finding := findingsResult.Finding
		result[name] = time.Since(time.Unix(finding.EventTime.Seconds, int64(finding.EventTime.Nanos)))
	}
	return result, nil
}
