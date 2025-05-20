# syntax=docker/dockerfile:1.2

# STAGE 1: build the executable
FROM golang:1.24-alpine AS builder

ARG BUILD_TARGET

RUN apk add --no-cache git openssh ca-certificates
WORKDIR /src

ARG TARGETARCH

# IF statement executed due to incosistent package names @ https://github.com/moparisthebest/static-curl/issues/8
RUN if [ "$TARGETARCH" = "arm64" ]; then \
        TARGETARCH=aarch64 ; \
        fi; \
    wget -O /usr/bin/curl https://github.com/moparisthebest/static-curl/releases/download/v8.0.1/curl-$TARGETARCH \
        && chmod +x /usr/bin/curl

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY internal internal
COPY $BUILD_TARGET $BUILD_TARGET

RUN --mount=type=cache,target=/root/.cache/go-build/ CGO_ENABLED=0 go build -o /app $BUILD_TARGET/main.go

# STAGE 2: build the container
FROM gcr.io/distroless/static:nonroot AS final

COPY --from=builder /usr/bin/curl /usr/bin/curl

USER 10000:10000

COPY --from=builder --chown=nonroot:nonroot /app /app

ENTRYPOINT ["/app"]
