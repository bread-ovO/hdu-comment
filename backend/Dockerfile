FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o server ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata bash sqlite-libs

COPY --from=builder /app/server ./server

# Initialize runtime directories (bind mount or volume in production if needed)
RUN mkdir -p /app/uploads /app/data

ENV APP_SERVER_PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]
