#!/usr/bin/env bash

DATESTR="$(date +%Y.%m.%d).${GITHUB_RUN_NUMBER}"
echo "Generated tag: ${DATESTR}"
echo "::set-output name=tag::${DATESTR}"