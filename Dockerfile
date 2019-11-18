# Build
FROM golang:1 as build

ADD . /app
WORKDIR /app

RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o cloudsec_metrics .

# Run
FROM alpine

RUN apk add --update ca-certificates && update-ca-certificates
RUN adduser -s /bin/false -S metrics

COPY --from=build /app/cloudsec_metrics /usr/bin

USER metrics

ENTRYPOINT ["/usr/bin/cloudsec_metrics"]