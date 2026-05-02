FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git build-base
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w -X main.Commit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}" -o /app/go-starter-kit .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /app/go-starter-kit /app/go-starter-kit
EXPOSE 8080
USER app
ENTRYPOINT ["/app/go-starter-kit", "serve"]
