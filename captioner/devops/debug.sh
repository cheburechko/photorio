#!/usr/bin/bash

REPO='cr.yandex/crpl95bufgfujr6vtbcp'
TAG='captioner:latest'

POD=$(kubectl get pods -l helm.sh/chart=captioner-0.1.0  -o jsonpath="{.items[0].metadata.name}")

kubectl exec --tty -i $POD -- bash
