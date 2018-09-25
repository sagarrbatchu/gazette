#!/usr/bin/env bash

# Parse explicit TAG option.
usage() { echo "Usage: $0 [ -t image-tag ]" 1>&2; exit 1; }

TAG="latest"

while getopts ":t:" opt; do
  case "${opt}" in
    t)   TAG=${OPTARG} ;;
    \? ) usage ;;
  esac
done

docker push liveramp/gazette:latest
docker push liveramp/gazette-example-stream-sum:latest
docker push liveramp/gazette-example-word-count:latest

if [ "$TAG" != "latest" ]; then
  docker tag liveramp/gazette:latest                    liveramp/gazette:${TAG}
  docker tag liveramp/gazette-example-stream-sum:latest liveramp/gazette-example-stream-sum:${TAG}
  docker tag liveramp/gazette-example-word-count:latest liveramp/gazette-example-word-count:${TAG}

  docker push liveramp/gazette:${TAG}
  docker push liveramp/gazette-example-stream-sum:${TAG}
  docker push liveramp/gazette-example-word-count:${TAG}
fi

