# syntax=docker/dockerfile:1
FROM golang:1.21-alpine as build
WORKDIR /app

COPY . .
RUN go build -mod vendor -o ./playground -buildvcs=false

FROM alpine:3
WORKDIR /app

COPY --from=build /app/assets ./assets
COPY --from=build /app/views ./views

COPY --from=build /app/playground /usr/bin/playground

# debugging
EXPOSE 9081

ENTRYPOINT ["playground"]
