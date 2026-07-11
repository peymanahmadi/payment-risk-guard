# --- build stage ---
FROM golang:1.23-bookworm AS build
WORKDIR /src

ENV GOPROXY=https://proxy.golang.com.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/service ./cmd/service

# --- runtime stage ---
FROM alpine:latest AS service
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/service /service
ENTRYPOINT ["/service"]