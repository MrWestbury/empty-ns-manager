#!/usr/bin/env bash
version="0.0.19"
docker build -t mortbury.azurecr.io/empty-ns-controller:${version} --build-arg "APPVERSION=${version}"  .