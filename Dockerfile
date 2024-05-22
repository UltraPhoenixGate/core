FROM golang:1.22-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
COPY . .
RUN go build -o bin/ultraphx-core -ldflags "-s -w" cmd/core/main.go

FROM alpine:3.12
WORKDIR /app
# Install ffmpeg
RUN apk add --no-cache ffmpeg

COPY --from=builder /app/bin/ultraphx-core /app/ultraphx-core

EXPOSE 8080
CMD ["/app/ultraphx-core"]
