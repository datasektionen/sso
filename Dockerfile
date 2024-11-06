ARG GO_VERSION=1.23
ARG ALPINE_VERSION=3.20

FROM docker.io/golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build

WORKDIR /src

COPY go.* ./

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    go mod download -x

COPY cmd cmd
COPY handlers handlers
COPY models models
COPY pkg pkg
COPY service service
COPY templates templates

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    CGO_ENABLED=0 go build -o /bin/server ./cmd/web

FROM docker.io/alpine:${ALPINE_VERSION}

ARG UID=10001
RUN adduser --disabled-password --gecos "" --home /nonexistent --shell "/sbin/nologin" \
    --no-create-home --uid "${UID}" user
USER user

COPY --from=build /bin/server /bin/

ENTRYPOINT ["/bin/server"]
