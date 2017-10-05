#!/bin/sh
set -x

docker build -t appscode/goldfish:canary .
docker push appscode/goldfish:canary
