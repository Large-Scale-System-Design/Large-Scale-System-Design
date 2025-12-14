# syntax=docker/dockerfile:1.7

FROM golang:1.22 AS build

ARG GOPROXY=direct
ARG GOSUMDB=off

ENV GOPROXY=$GOPROXY
ENV GOSUMDB=$GOSUMDB

WORKDIR /src

# Copy only module files first
COPY go.mod go.sum ./

# Download what we can (might be empty at this stage)
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source then make module graph consistent
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod tidy

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN CGO_ENABLED=0 GOOS=linux \
    go build -mod=mod -trimpath -o /out/urlshortener ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/urlshortener /app/urlshortener
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/urlshortener"]
