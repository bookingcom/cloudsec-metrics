# Cloud Security metrics collector [![Build Status](https://github.com/bookingcom/cloudsec-metrics/workflows/build/badge.svg)](https://github.com/bookingcom/cloudsec-metrics/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/bookingcom/cloudsec-metrics)](https://goreportcard.com/report/github.com/bookingcom/cloudsec-metrics)

Security-related metrics collector and exporter.

## How to run

```console
git clone git@github.com:bookingcom/cloudsec-metrics.git
cd cloudsec-metrics
# build a docker image with the application
docker-compose build metrics
docker-compose run metrics --help
```

### Parameters

| Command line            | Environment             | Default                  | Description                           |
| ----------------------- | ----------------------- | ------------------------ | ------------------------------------- |
| prisma_api_url          | PRISMA_API_URL          | https://api.eu.prismacloud.io | Prisma API key                   |
| prisma_api_key          | PRISMA_API_KEY          |                          | Prisma API key                        |
| prisma_api_password     | PRISMA_API_PASSWORD     |                          | Prisma API password                   |
| scc_org_id              | SCC_ORG_ID              |                          | Google SCC numeric organisation ID    |
| scc_sources_regex       | SCC_SOURCES_REGEX       | `.`                      | Google SCC sources Display Name filter regexp |
| graphite_host           | GRAPHITE_HOST           |                          | Graphite hostname                     |
| graphite_port           | GRAPHITE_PORT           | `2003`                   | Graphite port                         |
| graphite_prefix         | GRAPHITE_PREFIX         |                          | Global Graphite metrics prefix, applied to everything |
| compliance_prefix       | COMPLIANCE_PREFIX       | `compliance.`            | Graphite compliance metrics prefix    |
| dbg                     | DEBUG                   | `false`                  | debug mode                            |

## Overview

Collected metrics list:

- [Palo Alto Networks Prisma](https://www.paloaltonetworks.com/cloud-security):
  - assets compliance information per security standard
  - API health status ([SLA](https://www.paloaltonetworks.com/resources/datasheets/prisma-public-cloud-service-level-agreement))
- [Google Security Command Center](https://cloud.google.com/security-command-center/):
  - [health status](https://status.cloud.google.com/)
  - newest event update time per source (for monitoring [Forseti](https://forsetisecurity.org/) alerting delay).
  In order to collect this data, you need to specify `scc_org_id` and 
  have [proper credentials](https://cloud.google.com/docs/authentication/production) set up.

Supported exporters list:

- [Graphite](https://graphiteapp.org/)

## Acknowledgment

This software was originally developed at [Booking.com](http://www.booking.com).
With approval from [Booking.com](http://www.booking.com), this software was released
as Open Source, for which the authors would like to express their gratitude.
