#!/bin/bash -x

if [ "$TRAVIS_PULL_REQUEST" != "false" ]; then
    bash tools/build_container.sh
fi