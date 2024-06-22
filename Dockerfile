# Copyright (c) karl-cardenas-coding
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.22.4-alpine3.20 as builder

LABEL org.opencontainers.image.source="https://github.com/karl-cardenas-coding/mywhoop"
LABEL org.opencontainers.image.description "A tool for gathering and retaining your own Whoop data."

ARG VERSION

ADD ./ /source
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates && \
cd /source && \
adduser -H -u 1002 -D appuser appuser && \
go build -ldflags="-X 'github.com/karl-cardenas-coding/mywhoop/cmd.VersionString=${VERSION}'" -o whoop -v


FROM scratch

COPY --from=builder /source/whoop /go/bin/whoop
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group


USER appuser:appuser


ENTRYPOINT ["/go/bin/whoop"]

