FROM golang:1.16.5-alpine3.13
RUN apk add build-base
RUN mkdir -p /se5-websocket
WORKDIR /se5-websocket
COPY . .
RUN go mod download
RUN go build -o app
EXPOSE 8090
ENTRYPOINT ["./app"]