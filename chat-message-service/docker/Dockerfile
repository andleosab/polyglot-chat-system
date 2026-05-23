# ---- Builder ----
FROM golang:1.25.5-alpine AS build
WORKDIR /app
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

# golang-migrate CLI is shipped inside the runtime image so the same image can
# both serve traffic (default ENTRYPOINT) and run as a migration init-container.
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o main ./cmd/server

# ---- Runtime ----
FROM alpine:3.23
WORKDIR /app

RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S app -G app

COPY --from=build --chown=app:app /app/main /app/main
COPY --from=build --chown=app:app /go/bin/migrate /app/migrate
COPY --chown=app:app internal/db/migrations /app/migrations
USER app

ENV PORT=8080
EXPOSE 8080 9090

ENTRYPOINT ["/app/main"]
