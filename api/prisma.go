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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Prisma is an object to make API calls to Palo Alto Prisma,
// authorized requests if proper Token is set
type Prisma struct {
	Token  string `json:"token"`
	APIUrl string
}

// ComplianceInfo store assets compliance information for single policy
type ComplianceInfo struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	PoliciesCount     int    `json:"policiesAssignedCount"`
	PassedAssetsCount int    `json:"resourcesPassed"`
	FailedAssetsCount int    `json:"resourcesFailed"`
	TotalAssetsCount  int
}

// NewPrisma tries to log to Prisma in with given credentials and returns pointer to Prisma object on success, thread-safe
// https://api.docs.prismacloud.io/reference#app-login
func NewPrisma(username, password, apiURL string) (*Prisma, error) {
	loginData := map[string]string{"username": username, "password": password}
	jsonValue, err := json.Marshal(loginData)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling login data")
	}
	p := &Prisma{APIUrl: apiURL}
	data, err := p.doAPIRequest("POST", "/login", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, errors.Wrapf(err, "error logging in with user %q", username)
	}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, errors.Wrap(err, "error obtaining token")
	}
	return p, nil
}

// GatherComplianceInfo get assets compliance information for last day, thread-safe
// https://api.docs.prismacloud.io/reference#get-compliance-dashboard-list
func (p *Prisma) GatherComplianceInfo() ([]ComplianceInfo, error) {
	data, err := p.doAPIRequest("GET", "/compliance/dashboard?timeType=to_now&timeUnit=day", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting assets information")
	}

	var c []ComplianceInfo
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling assets information")
	}

	for i := range c {
		c[i].TotalAssetsCount = c[i].FailedAssetsCount + c[i].PassedAssetsCount
	}

	return c, nil
}

// GetAPIHealthStatus gets Prisma API health information and returns 1 on healthy response, 0 otherwise
// https://api.docs.prismacloud.io/reference#health-check
func (p *Prisma) GetAPIHealthStatus() int {
	if _, err := p.doAPIRequest("GET", "/check", nil); err != nil {
		return 0
	}
	return 1
}

// doAPIRequest does request to API with specified method and returns response body on success, thread-safe
func (p *Prisma) doAPIRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, p.APIUrl+url, body)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-redlock-auth", p.Token)
	httpClient := http.Client{Timeout: time.Second * 5}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error making request")
	}
	data, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading response body, response body: %q", data)
	}
	switch response.StatusCode {
	case http.StatusOK:
		return data, nil
	case http.StatusUnauthorized:
		// TODO handle token refresh in case it will be long-running
		return nil, errors.Errorf("authentication error on request, response body: %q", data)
	case http.StatusBadRequest:
		return nil, errors.Errorf("bad request parameters, check your request body, response body: %q", data)
	case http.StatusInternalServerError:
		return nil, errors.Errorf("server internal error during request processing, response body: %q", data)
	default:
		return nil, errors.Errorf("%v, response body: %q", response.Status, data)
	}
}
