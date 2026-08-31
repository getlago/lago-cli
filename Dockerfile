# syntax=docker/dockerfile:1.7
FROM golang:1.25.13-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X github.com/getlago/lago-cli/internal/cli.version=${VERSION}" -o /out/lago ./cmd/lago

FROM alpine:3.22 AS certificates
RUN apk add --no-cache ca-certificates

FROM scratch
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/lago /lago
USER 65532:65532
ENTRYPOINT ["/lago"]
