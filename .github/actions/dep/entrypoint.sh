#!/bin/bash

if [[ ! -z "$SSH_PRIVATE_KEY" ]]; then
  mkdir -p /ssh
  echo "$SSH_PRIVATE_KEY" > /ssh/id_rsa
  chmod 600 /ssh/id_rsa
  eval "$(ssh-agent -s)" > /dev/null 2>&1
  ssh-add /ssh/id_rsa > /dev/null 2>&1
fi

exec /go/bin/dep "$@"