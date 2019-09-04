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
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type googleStatusEntry struct {
	Service     string    `json:"service_key"`
	Description string    `json:"external_desc"`
	StartDate   time.Time `json:"begin"`
	EndDate     time.Time `json:"end"`
}

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
	data, err := ioutil.ReadAll(response.Body)
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
