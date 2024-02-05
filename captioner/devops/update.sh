#!/usr/bin/bash -x

TAG='captioner:latest'

minikube image build -t $TAG .. -f build/Dockerfile

helm uninstall captioner
helm install captioner ../helm
