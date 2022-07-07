FROM golang:1.18-alpine as build
WORKDIR /opt/build

COPY server server
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN go build -o server-executable main.go

FROM alpine
WORKDIR /opt/app

COPY --from=build /opt/build/server-executable .
COPY config.toml .

ENTRYPOINT ["/opt/app/server-executable"]