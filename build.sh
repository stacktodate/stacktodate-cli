#!/bin/bash

# Build script for stacktodate across multiple platforms
#
# Note: This script is provided for local testing and development builds.
# For official releases, version tags are used to automatically trigger GoReleaser
# via GitHub Actions, which builds and publishes binaries across all platforms.
# See README.md "Creating a Release" section for the release process.

set -e

BINARY_NAME="stacktodate"
PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

OUTPUT_DIR="dist"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "Building $BINARY_NAME for multiple platforms..."

for platform in "${PLATFORMS[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "$platform"

  # Determine output filename
  if [ "$GOOS" = "windows" ]; then
    OUTPUT="$OUTPUT_DIR/${BINARY_NAME}_${GOOS}_${GOARCH}.exe"
  else
    OUTPUT="$OUTPUT_DIR/${BINARY_NAME}_${GOOS}_${GOARCH}"
  fi

  echo "Building for $GOOS/$GOARCH -> $OUTPUT"
  GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT"
done

echo ""
echo "Build complete! Binaries are in the $OUTPUT_DIR directory:"
ls -lh "$OUTPUT_DIR"
