FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates=20260611-r0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api      ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrator ./cmd/migrator

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates=20260413-r0 curl=8.14.1-r2 tzdata=2026b-r0 \
 && adduser -D -u 10001 app
WORKDIR /app
COPY config/config.yaml ./config/config.yaml
COPY migrations ./migrations
COPY api/openapi.yaml ./openapi.yaml
COPY --from=build /out/api /out/migrator ./
USER 10001
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
  CMD ["curl", "-fsS", "http://localhost:8080/health/live"]
ENTRYPOINT ["./api"]
