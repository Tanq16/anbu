FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine3.24 AS builder

WORKDIR /app

RUN apk add --no-cache git curl make uv

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH VERSION=dev-build

RUN make assets && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
      -ldflags="-s -w -X 'github.com/tanq16/anbu/cmd.AppVersion=${VERSION}'" \
      -o /app/anbu .

FROM alpine:3.24.2

RUN apk add --no-cache ca-certificates tzdata aws-cli && \
    addgroup -g 10001 -S app && \
    adduser -u 10001 -S -G app app

WORKDIR /app
COPY --from=builder --chown=10001:10001 /app/anbu .

RUN mkdir -p /data && chown 10001:10001 /data
VOLUME ["/data"]

USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["./anbu"]
CMD ["serve", "-d", "/data", "-H", "0.0.0.0"]
