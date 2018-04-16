#!/bin/sh
set -x

docker build -t appscode/goldfish:0.7.0 .
docker push appscode/goldfish:0.7.0
