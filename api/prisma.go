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
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Prisma is an object to make API calls to Palo Alto Prisma,
// authorized requests if proper Token is set
type Prisma struct {
	Login          string
	Password       string
	Token          string `json:"token"`
	APIUrl         string
	tokenRenewTime time.Time
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

// prismRenewTimeout defines how often auth token is renewed,
// after 10 minutes it gets invalidated and new complete login is required
const prismRenewTimeout = time.Minute * 3

// GatherComplianceInfo get assets compliance information for last day
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

// doAPIRequest does request to API with specified method and returns response body on success
func (p *Prisma) doAPIRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, p.APIUrl+url, body)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}
	req.Header.Set("Content-Type", "application/json")
	if time.Since(p.tokenRenewTime) > prismRenewTimeout {
		if err := p.authenticate(); err != nil {
			return nil, errors.Wrap(err, "error getting auth token")
		}
	}
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
		return nil, errors.Errorf("authentication error on request, response body: %q", data)
	case http.StatusBadRequest:
		return nil, errors.Errorf("bad request parameters, check your request body, response body: %q", data)
	case http.StatusInternalServerError:
		return nil, errors.Errorf("server internal error during request processing, response body: %q", data)
	default:
		return nil, errors.Errorf("%v, response body: %q", response.Status, data)
	}
}

// authenticate gets or renews the API authentication token
// https://api.docs.prismacloud.io/reference#login
func (p *Prisma) authenticate() error {
	p.tokenRenewTime = time.Now()
	switch p.Token {
	case "":
		// no token set yet, first login
		loginData := map[string]string{"username": p.Login, "password": p.Password}
		jsonValue, err := json.Marshal(loginData)
		if err != nil {
			return errors.Wrap(err, "error marshaling login data")
		}
		data, err := p.doAPIRequest("POST", "/login", bytes.NewBuffer(jsonValue))
		if err != nil {
			return errors.Wrapf(err, "error logging in with user %q", p.Login)
		}
		if err := json.Unmarshal(data, p); err != nil {
			return errors.Wrap(err, "error obtaining token from login response")
		}
	default:
		// token is set and we will try to renew it
		data, err := p.doAPIRequest("GET", "/auth_token/extend", nil)
		if err != nil {
			log.Printf("[INFO] Error extending token, will re-login, %v", err)
			p.Token = ""
			return p.authenticate()
		}
		if err := json.Unmarshal(data, p); err != nil {
			log.Printf("[INFO] Error obtaining token from extend token response, will re-login, %v", err)
			p.Token = ""
			return p.authenticate()
		}
	}
	return nil
}
