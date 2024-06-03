ARG GO_VERSION=1.22
ARG NODE_VERSION=22.2
ARG ALPINE_VERSION=3.20

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build-server

WORKDIR /src

COPY go.* ./

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    go mod download -x

COPY pkg pkg
COPY services services
COPY cmd cmd

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    CGO_ENABLED=0 go build -o /bin/server ./cmd/web

FROM node:${NODE_VERSION}-alpine${ALPINE_VERSION} AS build-islands

WORKDIR /src

RUN corepack enable

COPY package.json pnpm-lock.yaml ./

RUN pnpm i

COPY rollup.config.mjs tsconfig.json ./
COPY islands/ islands/

RUN pnpm build

FROM alpine:${ALPINE_VERSION}

ARG UID=10001
RUN adduser --disabled-password --gecos "" --home /nonexistent --shell "/sbin/nologin" \
    --no-create-home --uid "${UID}" user
USER user

COPY --from=build-server /bin/server /bin/
COPY --from=build-islands /src/dist /dist/

ENV DIST_DIR=/dist

ENTRYPOINT ["/bin/server"]
