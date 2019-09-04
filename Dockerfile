# Lint check
FROM golangci/golangci-lint

WORKDIR /test
ADD . .

RUN golangci-lint run --disable-all --deadline=300s --enable=vet --enable=vetshadow --enable=golint \
    --enable=staticcheck --enable=ineffassign --enable=goconst --enable=errcheck --enable=unconvert \
    --enable=deadcode --enable=gosimple ./...

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