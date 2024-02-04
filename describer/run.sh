#!/usr/bin/bash

REPO='cr.yandex/crpl95bufgfujr6vtbcp'
TAG='describer:latest'

docker build . -t $TAG
docker run --rm -it --init --gpus=all $TAG
