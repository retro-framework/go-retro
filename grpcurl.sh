#!/usr/bin/env bash
set -o nounset
set -o errexit

IMAGE="quay.io/datawire/grpcurl"
VERSION="latest"

docker run \
  --interactive \
  --tty \
  --rm \
  --network host \
  --volume $(pwd):/home/user/work:ro \
  --workdir /home/user/work \
  -e "COMMAND=${COMMAND}" \
  -e HOST_USER_ID=$(id -u) \
  "$IMAGE:$VERSION" "$@"
