#!/bin/bash
# build.sh - Flexible cross-platform build script for gitdig
set -e
VERSION=$(git describe --tags 2>/dev/null || echo "dev")
echo "Building gitdig version: $VERSION"
BUILD_DIR="./build"
mkdir -p "$BUILD_DIR"
echo "Cleaning previous builds..."
rm -rf "$BUILD_DIR"/*

# Default full platform list
ALL_PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "darwin/amd64"
    "darwin/arm64"
)

# Default output format
OUTPUT_FORMAT="all"

# Parse arguments
PLATFORMS=()
while [[ $# -gt 0 ]]; do
    case $1 in
        --format=*)
            OUTPUT_FORMAT="${1#*=}"
            shift
            ;;
        --format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        all)
            PLATFORMS=("${ALL_PLATFORMS[@]}")
            shift
            ;;
        *)
            PLATFORMS+=("$1")
            shift
            ;;
    esac
done

# Determine platforms to build for
if [ ${#PLATFORMS[@]} -eq 0 ]; then
    # No platform args: only build for current platform
    current_platform="$(go env GOOS)/$(go env GOARCH)"
    PLATFORMS=("$current_platform")
    echo "No platform specified. Building for current platform: $current_platform"
else
    echo "Building for specified platforms: ${PLATFORMS[*]}"
fi

# Validate output format
case "$OUTPUT_FORMAT" in
    binary|bin)
        echo "Output format: binary only"
        OUTPUT_FORMAT="binary"
        ;;
    archive|tar|zip)
        echo "Output format: archives only (tar.gz/zip)"
        OUTPUT_FORMAT="archive"
        ;;
    all)
        echo "Output format: both binary and archives"
        OUTPUT_FORMAT="all"
        ;;
    *)
        echo "Invalid output format: $OUTPUT_FORMAT. Using 'all' as default."
        OUTPUT_FORMAT="all"
        ;;
esac

# Loop through platforms
for platform in "${PLATFORMS[@]}"; do
    IFS="/" read -r GOOS GOARCH <<< "$platform"
    output_name="gitdig-v$VERSION-${GOOS}-${GOARCH}"
    bin_name="$output_name"
    [ "$GOOS" = "windows" ] && bin_name+='.exe'

    echo "Building for $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-X 'main.Version=$VERSION'" \
        -o "$BUILD_DIR/$bin_name" \
        main.go

    if [ $? -eq 0 ]; then
        echo "✅ Successfully built $bin_name"

        # Handle output formats
        if [[ "$OUTPUT_FORMAT" == "archive" || "$OUTPUT_FORMAT" == "all" ]]; then
            echo "Creating archive..."
            if [ "$GOOS" = "windows" ]; then
                (cd "$BUILD_DIR" && zip -q "$output_name.zip" "$bin_name")
                echo "✅ Created $output_name.zip"
            else
                tar -czf "$BUILD_DIR/$output_name.tar.gz" -C "$BUILD_DIR" "$bin_name"
                echo "✅ Created $output_name.tar.gz"
            fi
        fi

        # Clean up binary if not requested
        if [[ "$OUTPUT_FORMAT" == "archive" ]]; then
            rm "$BUILD_DIR/$bin_name"
        fi
    else
        echo "❌ Failed to build for $GOOS/$GOARCH"
    fi
done

echo ""
echo "Build summary:"
ls -la "$BUILD_DIR"
echo ""
echo "🎉 Build process completed!"