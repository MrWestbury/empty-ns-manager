#!/usr/bin/env bash

DATESTR="$(date +%Y.%m.%d).${GITHUB_RUN_ID}"
echo "Generated tag: ${DATESTR}"
echo "::set-output name=tag::${DATESTR}"