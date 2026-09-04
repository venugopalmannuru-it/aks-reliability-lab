# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26.8

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine3.24 AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o /out/platform-api ./cmd/api

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/platform-api /platform-api

USER 65532:65532
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 CMD ["/platform-api", "-healthcheck"]
ENTRYPOINT ["/platform-api"]
