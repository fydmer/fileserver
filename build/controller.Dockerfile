FROM golang:1.23.2-alpine3.20 AS build

WORKDIR /build
COPY . ./

RUN go mod tidy -v \
    && CGO_ENABLED=0 GOOS=linux go build -o /tmp/controller ./cmd/controller/main.go

FROM alpine:3.20

WORKDIR /app
COPY --from=build /tmp/controller ./controller

CMD ["/app/controller"]