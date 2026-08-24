#!/bin/sh
set -eu
docker build -f benzhi.Dockerfile -t mineair:local .
