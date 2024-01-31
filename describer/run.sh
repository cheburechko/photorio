#!/usr/bin/bash

TAG='cr.yandex/crpl95bufgfujr6vtbcp/describer:latest'

docker build . -t $TAG
docker run --rm -it --init --gpus=all $TAG
