#!/bin/sh
if ${GITHUB_TOKEN+false}; then
  echo GITHUB_TOKEN is unset, please follow README
  exit 1
fi
#git config --global --add url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/manabie-com".insteadOf "https://github.com/manabie-com"
modd