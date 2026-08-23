ARG GO_BUILD_ARGS="-trimpath -ldflags -s -w"

FROM golang:1.25-alpine AS builder

WORKDIR /app
ENV CGO_ENABLED=0
ENV GOOS=linux

COPY go.mod go.sum ./

RUN go mod download

COPY . .
# One binary with subcommands (server|setup-index|recreate-index|
# incremental-index|crawl) instead of five ~50MB binaries, so image
# updates transfer a single small layer.
RUN go build ${GO_BUILD_ARGS} -o bin/rm-search .

FROM alpine:3.20

# tzdata lets the binary resolve TZ names like Asia/Shanghai; without it
# the Go runtime silently falls back to UTC even when /etc/localtime is
# mounted. ca-certificates-bundle is already in the alpine base image.
RUN apk add --no-cache tzdata

WORKDIR /root/

COPY --from=builder /app/bin/rm-search /usr/local/bin/rm-search

# Byte-sized compatibility wrappers for the former standalone tools.
RUN for tool in setup-index recreate-index incremental-index crawl; do \
        printf '#!/bin/sh\nexec /usr/local/bin/rm-search %s "$@"\n' "$tool" \
            > "/usr/local/bin/$tool" && chmod +x "/usr/local/bin/$tool"; \
    done

EXPOSE 8080

ENV GIN_MODE=release
ENV PROD=true

# Timezone via mounted host files (see deploy/docker-compose.yml); TZ env
# overrides when set.
ENV TZ=Asia/Shanghai

ENTRYPOINT ["/usr/local/bin/rm-search"]
