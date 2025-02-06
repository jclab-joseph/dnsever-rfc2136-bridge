FROM golang:1.22-alpine3.20 as builder

RUN mkdir -p /build/
COPY . /build/

RUN cd /build/ && \
    ls -al && \
    CGO_ENABLED=0 go build -o /build/dnsever-rfc2136-bridge ./cmd/server

FROM alpine:3.20
COPY --from=builder /build/dnsever-rfc2136-bridge /dnsever-rfc2136-bridge

CMD /dnsever-rfc2136-bridge
