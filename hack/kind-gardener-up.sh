#!/bin/bash

cd "$1"
chmod +rw . -R
chmod +rwx ./hack -R
make kind-up
make gardener-up
