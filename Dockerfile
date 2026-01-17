ARG GO_BUILD_ARGS="-trimpath -ldflags -s -w"

FROM node:20-slim AS frontend-builder

WORKDIR /app
RUN npm install -g pnpm

COPY ./frontend/package.json ./frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY ./frontend .

RUN pnpm build

FROM golang:1.23-alpine as builder

WORKDIR /app
ENV CGO_ENABLED=0
ENV GOOS=linux

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build ${GO_BUILD_ARGS} -o bin/recreate-index ./cmd/recreate-index
RUN go build ${GO_BUILD_ARGS} -o bin/incremental-index ./cmd/incremental-index

COPY --from=frontend-builder /app/dist ./server/handler/frontend
RUN go build ${GO_BUILD_ARGS} -o bin/rm-search .

FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /app/bin/recreate-index /usr/local/bin/recreate-index
COPY --from=builder /app/bin/incremental-index /usr/local/bin/incremental-index
COPY --from=builder /app/bin/rm-search /usr/local/bin/rm-search

EXPOSE 8080

ENV GIN_MODE=release
ENV PROD=true

ENTRYPOINT ["/usr/local/bin/rm-search"]