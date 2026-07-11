# --- build stage ---
FROM golang:1.23-bookworm AS build
WORKDIR /src

# Set Go proxy to use multiple mirrors
ENV GOPROXY=https://proxy.golang.com.cn,direct
# OR use direct downloads without proxy
# ENV GOPROXY=direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/service ./cmd/service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/simulator ./cmd/simulator

# --- runtime stage ---
FROM alpine:latest AS service
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/service /service
ENTRYPOINT ["/service"]

FROM alpine:latest AS simulator
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/simulator /simulator
ENTRYPOINT ["/simulator"]