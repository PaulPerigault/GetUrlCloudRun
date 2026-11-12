# syntax=docker/dockerfile:1
# Multi-stage build: pure Go binary (CGO off) in a distroless image that runs
# as the numeric non-root user 65532. Base images are pinned by digest;
# Dependabot (docker ecosystem) proposes the updates.

FROM golang:1.27.4-alpine@sha256:0a5f7fe796599ff487605987432e7619ae45beeb2fba4c2ce0b907879932aebf AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY cmd ./cmd
COPY internal ./internal
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w \
        -X github.com/PaulPerigault/GetUrlCloudRun/internal/version.Version=${VERSION} \
        -X github.com/PaulPerigault/GetUrlCloudRun/internal/version.Commit=${COMMIT} \
        -X github.com/PaulPerigault/GetUrlCloudRun/internal/version.Date=${DATE}" \
      -o /out/geturlcloudrun ./cmd/geturlcloudrun

FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
LABEL org.opencontainers.image.title="geturlcloudrun" \
      org.opencontainers.image.description="Read-only HTTP API that returns the URLs of Cloud Run services" \
      org.opencontainers.image.source="https://github.com/PaulPerigault/GetUrlCloudRun" \
      org.opencontainers.image.licenses="MIT"
COPY --from=build /out/geturlcloudrun /geturlcloudrun
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/geturlcloudrun"]
CMD ["serve"]
