#!/bin/bash
set -e

# Ensure we are in the script's directory
cd "$(dirname "$0")"

BUILDER_TAG="v10.0.0"
SCRIPT_URL="https://raw.githubusercontent.com/multiversx/mx-sdk-rust-contract-builder/${BUILDER_TAG}/build_with_docker.py"
SCRIPT_NAME="build_with_docker.py"

# Download the build script
echo "Downloading build script from ${SCRIPT_URL}..."
curl -s -o "${SCRIPT_NAME}" "${SCRIPT_URL}"

# Clean and create output directory
rm -rf output
mkdir -p output

# Run the build
echo "Starting reproducible build using Docker image multiversx/sdk-rust-contract-builder:${BUILDER_TAG}..."
python3 "${SCRIPT_NAME}" \
  --image="multiversx/sdk-rust-contract-builder:${BUILDER_TAG}" \
  --project="$(pwd)" \
  --output="$(pwd)/output" \
  --no-docker-interactive \
  --contract=requests

echo "Build complete. Artifacts are in $(pwd)/output"
