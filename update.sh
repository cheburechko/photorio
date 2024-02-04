#!/usr/bin/bash -x

TAG='describer:latest'

minikube image build -t $TAG ./describer

helm uninstall describer
helm install describer ./helm/describer
