FROM golang:1.14-alpine AS builder
RUN mkdir -p /build
WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o app
CMD ["./app"]