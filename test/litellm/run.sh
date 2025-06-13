#!/usr/bin/env bash

if [ -z "$OPENAI_API_KEY" ]; then
  echo "Please set the OPENAI_API_KEY environment variable."
  exit 1
fi

docker run -p 4000:4000 \
  -v "$(pwd)/config.yaml:/app/config.yaml" \
  -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
  -e LITELLM_API_KEYS="my_api_key" \
  --platform "linux/amd64" \
  --entrypoint /bin/bash \
  ghcr.io/berriai/litellm:main \
  -c "litellm --config /app/config.yaml --port 4000"