# ---- Builder ----
FROM golang:1.25.5-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o main ./cmd/server

# ---- Runtime ----
FROM alpine:3.23
WORKDIR /app

RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S app -G app

COPY --from=build --chown=app:app /app/main /app/main
USER app

ENV PORT=8080
EXPOSE 8080 9090

ENTRYPOINT ["/app/main"]
