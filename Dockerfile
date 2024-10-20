# Copyright (c) karl-cardenas-coding
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.23.2-alpine3.20 as builder

LABEL org.opencontainers.image.source="https://github.com/karl-cardenas-coding/mywhoop"
LABEL org.opencontainers.image.description="A tool for gathering and retaining your own Whoop data."

ARG VERSION

RUN apk update && \
    apk add --no-cache ca-certificates tzdata xdg-utils && \
    update-ca-certificates && \
    adduser -H -u 1002 -D appuser appuser

ADD ./ /source
WORKDIR /source
RUN go build -ldflags="-X 'github.com/karl-cardenas-coding/mywhoop/cmd.VersionString=${VERSION}'" -o mywhoop -v && \
mkdir -p /app && chown -R appuser:appuser /app


FROM scratch

USER appuser:appuser

COPY --from=builder /source/mywhoop /bin/mywhoop
COPY --from=builder /usr/bin/xdg-open /usr/bin/xdg-open
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app /app

EXPOSE 8080


ENTRYPOINT ["/bin/mywhoop"]

