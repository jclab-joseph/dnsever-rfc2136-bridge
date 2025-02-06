FROM golang:1.23-alpine3.21 as builder

RUN mkdir -p /build/
COPY . /build/

RUN cd /build/ && \
    ls -al && \
    CGO_ENABLED=0 go build -o /build/dnsever-rfc2136-bridge ./cmd/server

FROM alpine:3.21
COPY --from=builder /build/dnsever-rfc2136-bridge /dnsever-rfc2136-bridge

CMD /dnsever-rfc2136-bridge
