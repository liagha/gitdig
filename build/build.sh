#!/bin/bash
# build.sh - Cross-platform build script for gitdig
# This script compiles gitdig for multiple platforms and architectures

set -e  # Exit immediately if a command exits with a non-zero status

# Define version from git tag or default to development
VERSION=$(git describe --tags 2>/dev/null || echo "dev")
echo "Building gitdig version: $VERSION"

# Define output directory
BUILD_DIR="./build"
mkdir -p "$BUILD_DIR"

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf "$BUILD_DIR"/*

# Define platforms to build for
platforms=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "darwin/amd64"
    "darwin/arm64"
)

# Build for each platform
for platform in "${platforms[@]}"
do
    # Split platform into OS and architecture
    IFS="/" read -r GOOS GOARCH <<< "$platform"

    # Set output name with version
    output_name="gitdig-v$VERSION-${GOOS}-${GOARCH}"

    # Add .exe extension for Windows
    if [ "$GOOS" = "windows" ]; then
        output_name+='.exe'
    fi

    echo "Building for $GOOS/$GOARCH..."

    # Build binary with version information
    env GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-X 'main.Version=$VERSION'" \
        -o "$BUILD_DIR/$output_name" \
        main.go

    # Check if build was successful
    if [ $? -eq 0 ]; then
        echo "✅ Successfully built $output_name"

        # Create compressed archive
        echo "Creating archive for $output_name..."
        if [ "$GOOS" = "windows" ]; then
            # Create ZIP archive for Windows
            (cd "$BUILD_DIR" && zip -q "$output_name.zip" "$output_name")
            rm "$BUILD_DIR/$output_name"
        else
            # Create tar.gz archive for Unix-like platforms
            tar -czf "$BUILD_DIR/$output_name.tar.gz" -C "$BUILD_DIR" "$output_name"
            rm "$BUILD_DIR/$output_name"
        fi
    else
        echo "❌ Failed to build for $GOOS/$GOARCH"
    fi
done

# Show summary
echo ""
echo "Build summary:"
ls -la "$BUILD_DIR"

echo ""
echo "🎉 Build process completed!"