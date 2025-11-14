#!/usr/bin/bash -x

docker build -f backend/build/worker/Dockerfile backend -t photorio/worker
docker build -f backend/build/api/Dockerfile backend -t photorio/api