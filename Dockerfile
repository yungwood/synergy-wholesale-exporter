# syntax=docker/dockerfile:1

# Stage 1: build golang binary
FROM golang:1.26-alpine as builder
ARG VERSION="unknown"
ARG REVISION="unknown"
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY *.go ./
COPY synergywholesaleapi/ ./synergywholesaleapi/
RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.version=${VERSION}' -X 'main.revision=${REVISION}'" -tags timetzdata -o /go/bin/synergy-wholesale-exporter .

# Stage 2: setup alpine base for building scratch image
FROM alpine:3.24.1 as base
RUN adduser -s /bin/true -u 1000 -D -h /app app && \
  sed -i -r "/^(app|root)/!d" /etc/group /etc/passwd && \
  sed -i -r 's#^(.*):[^:]*$#\1:/sbin/nologin#' /etc/passwd

# Stage 3: create final image from scratch
FROM scratch
ARG VERSION="unknown"
ARG REVISION="unknown"
LABEL org.opencontainers.image.title="synergy-wholesale-exporter" \
  org.opencontainers.image.description="Prometheus exporter for Synergy Wholesale domain metrics" \
  org.opencontainers.image.version="${VERSION}" \
  org.opencontainers.image.revision="${REVISION}" \
  org.opencontainers.image.source="https://github.com/yungwood/synergy-wholesale-exporter" \
  org.opencontainers.image.licenses="MIT"
WORKDIR /app
COPY --from=base /etc/passwd /etc/group /etc/
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/synergy-wholesale-exporter /usr/bin/synergy-wholesale-exporter
USER 1000:1000
EXPOSE 8080/tcp
ENTRYPOINT ["/usr/bin/synergy-wholesale-exporter"]
CMD []
