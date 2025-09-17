# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

# Build arguments for version and target platform
ARG VERSION=unknown
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build with version info and target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags="-X artship/internal/version.Version=${VERSION} -s -w" \
    -o app ./cmd/artship

# Runtime stage (scratch)
FROM scratch

LABEL org.opencontainers.image.source=https://github.com/ipaqsa/artship
LABEL org.opencontainers.image.description="A CLI tool to extract artifact from images"
LABEL org.opencontainers.image.licenses=MIT

COPY --from=builder /app/app /artship

ENTRYPOINT ["/artship"]