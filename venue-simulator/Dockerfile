FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /venue-simulator .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /venue-simulator /app/venue-simulator
COPY menu.json /app/menu.json
EXPOSE 8081
ENTRYPOINT ["/app/venue-simulator"]
