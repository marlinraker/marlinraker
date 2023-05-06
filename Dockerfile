FROM golang:1.20.3-alpine3.16 AS build

WORKDIR /build
COPY ./ /build
RUN go build -o marlinraker -ldflags "-s -w" src/main.go

FROM alpine:3.16

RUN adduser -h /marlinraker -G dialout -u 1001 -D marlinraker
WORKDIR /marlinraker
USER marlinraker

RUN mkdir marlinraker_files
COPY --from=build /build/marlinraker .

EXPOSE 7125
VOLUME ["/marlinraker/marlinraker_files"]
CMD ["./marlinraker"]