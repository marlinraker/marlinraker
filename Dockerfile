FROM golang:1.27-alpine3.24 AS build

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
ARG VERSION=0.0.0-dev
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X github.com/marlinraker/marlinraker/internal/constants.Version=${VERSION}" \
    -o marlinraker ./cmd/marlinraker

FROM alpine:3.24

RUN adduser -h /marlinraker -G dialout -u 1001 -D marlinraker
WORKDIR /marlinraker
USER marlinraker

RUN mkdir marlinraker_files
COPY --from=build /build/marlinraker .

EXPOSE 7125
VOLUME ["/marlinraker/marlinraker_files"]
CMD ["./marlinraker"]
