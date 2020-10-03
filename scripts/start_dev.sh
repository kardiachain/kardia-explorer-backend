#!/bin/sh
if ${GITHUB_TOKEN+false}; then
  echo GITHUB_TOKEN is unset, please follow README
  exit 1
fi
modd