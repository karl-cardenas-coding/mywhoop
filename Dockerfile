FROM golang:1.22.3-alpine3.19 as builder

LABEL org.opencontainers.image.source="https://github.com/karl-cardenas-coding/mywhoop"
LABEL org.opencontainers.image.description "A tool for gathering and retaining your own Whoop data."

ARG VERSION

ADD ./ /source
RUN cd /source && \
adduser -H -u 1002 -D appuser appuser && \
go build -ldflags="-X 'github.com/karl-cardenas-coding/mywhoop/cmd.VersionString=${VERSION}'" -o whoop -v


FROM scratch

COPY --from=builder /source/whoop /go/bin/whoop
ENTRYPOINT ["/go/bin/whoop"]

