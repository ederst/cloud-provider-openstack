#!/usr/bin/env bash

ARCH=amd64 \
GOOS=linux \
REGISTRY=docker.io/infonova \
VERSION=1.22.1-bp-fd2a9277-1 \
make image-openstack-cloud-controller-manager