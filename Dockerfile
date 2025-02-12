# Stage 1: build golang binary
FROM golang:1.24-alpine as builder
ARG VERSION="unknown"
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 go install -ldflags "-extldflags '-static' -X 'main.version=${VERSION}'" -tags timetzdata

# Stage 2: setup alpine base for building scratch image
FROM alpine:3.21.2 as base
RUN adduser -s /bin/true -u 1000 -D -h /app app && \
  sed -i -r "/^(app|root)/!d" /etc/group /etc/passwd && \
  sed -i -r 's#^(.*):[^:]*$#\1:/sbin/nologin#' /etc/passwd

# Stage 3: create final image from scratch
FROM scratch
WORKDIR /app
COPY --from=base /etc/passwd /etc/group /etc/shadow /etc/
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/synergy-wholesale-exporter /usr/bin/synergy-wholesale-exporter
USER app
EXPOSE 8080/tcp
ENTRYPOINT ["/usr/bin/synergy-wholesale-exporter"]
CMD []
