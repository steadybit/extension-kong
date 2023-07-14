#!/bin/sh -e

#
# Copyright 2023 steadybit GmbH. All rights reserved.
#

if ! getent passwd steadybit >/dev/null 2>&1; then
  useradd --system steadybit
  printf "created user: steadybit\n"
fi

if getent group docker >/dev/null 2>&1; then
  gpasswd -a steadybit docker
fi
