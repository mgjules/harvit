FROM golang:1.18-alpine AS builder

# Add git, curl and upx support
RUN apk add --no-cache git curl upx

WORKDIR /src

# Pull modules
COPY go.* ./
RUN go mod download

# Copy code into image
COPY . ./

# Build application for deployment
RUN CGO_ENABLED=0 go build -tags=jsoniter -trimpath -ldflags "-s -w" -o /tmp/harvit .

# Compress binary
RUN upx --best --lzma /tmp/harvit

# Create minimal image with just the application
# gcr.io/distroless/static is perfect for Go app that do not depend on libc
FROM gcr.io/distroless/static
COPY --from=builder /tmp/harvit /harvit
CMD ["/harvit", "harvest"]
