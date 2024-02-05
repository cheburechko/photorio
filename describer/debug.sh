#!/usr/bin/bash

REPO='cr.yandex/crpl95bufgfujr6vtbcp'
TAG='describer:latest'

POD=$(kubectl get pods -l helm.sh/chart=describer-0.1.0  -o jsonpath="{.items[0].metadata.name}")

kubectl exec --tty -i $POD -- bash
