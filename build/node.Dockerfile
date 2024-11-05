FROM golang:1.23.2-alpine3.20 AS build

WORKDIR /build
COPY . ./

RUN go mod tidy -v \
    && CGO_ENABLED=0 GOOS=linux  go build -o /tmp/node ./cmd/node/main.go

FROM alpine:3.20

WORKDIR /app
COPY --from=build /tmp/node ./node

CMD ["/app/node"]