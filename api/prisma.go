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
	"io"

	"github.com/paskal/go-prisma"
	"github.com/pkg/errors"
)

// Prisma contain credentials for API access
type Prisma struct {
	api apiCaller
}

type apiCaller interface {
	DoAPIRequest(method, url string, body io.Reader) ([]byte, error)
}

// ComplianceInfo store assets compliance information for single policy
type ComplianceInfo struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	PoliciesCount     int    `json:"assignedPolicies"`
	PassedAssetsCount int    `json:"passedResources"`
	FailedAssetsCount int    `json:"failedResources"`
	TotalAssetsCount  int    `json:"totalResources"`
}

// CompliancePosture stores overall posture statistics
// and is required to unwrap the nested JSON scheme
type CompliancePosture struct {
	ComplianceDetails []ComplianceInfo `json:"complianceDetails"`
}

// NewPrisma returns new Prisma client
func NewPrisma(username, password, apiURL string) *Prisma {
	p := Prisma{}
	p.api = prisma.NewClient(username, password, apiURL)
	return &p
}

// GatherComplianceInfo get assets compliance information for last day
// https://api.docs.prismacloud.io/reference#compliance-posture
func (p *Prisma) GatherComplianceInfo() ([]ComplianceInfo, error) {
	data, err := p.api.DoAPIRequest("GET", "/compliance/posture?timeType=to_now&timeUnit=day", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting assets information")
	}

	var posture CompliancePosture
	if err := json.Unmarshal(data, &posture); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling assets information")
	}

	return posture.ComplianceDetails, nil
}

// GetAPIHealthStatus gets Prisma API health information and returns 1 on healthy response, 0 otherwise
// https://api.docs.prismacloud.io/reference#health-check
func (p *Prisma) GetAPIHealthStatus() int {
	if _, err := p.api.DoAPIRequest("GET", "/check", nil); err != nil {
		return 0
	}
	return 1
}
