#!/usr/bin/env bash

set -euo pipefail

# Constants
readonly RUNNER_USER_UID=1001
readonly DOCKER_GROUP_GID=121

# Update apt and install initial dependencies
install_dependencies() {
  apt-get update -y

  apt-get install -y software-properties-common

  add-apt-repository -y ppa:git-core/ppa

  apt-get update -y

  apt-get install -y --no-install-recommends curl ca-certificates jq sudo unzip zip

  rm -rf /var/lib/apt/lists/*
}

# Create user for runner and add it to sudo and docker groups
create_user() {
  # Create user for runner
  adduser --disabled-password --gecos "" --uid $RUNNER_USER_UID runner

  # Add new docker group with same GID as host docker group
  groupadd docker --gid $DOCKER_GROUP_GID

  # Add runner to sudo and docker groups
  usermod -aG sudo runner
  usermod -aG docker runner

  # Allow sudo without password
  echo "%sudo   ALL=(ALL:ALL) NOPASSWD:ALL" >/etc/sudoers

  # Allow sudo to set DEBIAN_FRONTEND
  echo "Defaults env_keep += \"DEBIAN_FRONTEND\"" >>/etc/sudoers
}

create_known_directories() {
  # Create actions directory and set owner to runner home directory. This directory used by actions.
  mkdir -p /home/actions
  chown -R runner:runner /home/actions

  # Temp directory is defined in RUNNER_TEMP environment variable. This directory will be used to store temporary files.
  # However, when we use dagger to mount directory to container, the directory will be owned by root. To workaround this
  # issue, we create the directory and set owner to runner.
  mkdir -p /home/runner/_temp
  chown -R runner:runner /home/runner/_temp
}

main() {
  install_dependencies
  create_user
  create_known_directories
}

main "$@"
