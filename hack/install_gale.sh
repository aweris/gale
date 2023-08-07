#!/usr/bin/env bash

RELEASE=$(curl -s "https://api.github.com/repos/aweris/gale/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
VERSION="${VERSION:-${RELEASE}}"
INSTALL_DIR="${INSTALL_DIR:-.}"

function install_gale() {
    local version=$1
    local os=$2
    local arch=$3
    local file_name=gale_${version}_${os}_${arch}.tar.gz
    local download_url=https://github.com/aweris/gale/releases/download/${version}/${file_name}

    echo "Downloading ${download_url}"

    curl -sL "${download_url}" | tar xz -C "${INSTALL_DIR}"

    echo "Installed gale ${version} to ${INSTALL_DIR}"

    "${INSTALL_DIR}/gale" version

    echo "Done."
}

function main() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')

    local arch
    arch=$(uname -m | tr '[:upper:]' '[:lower:]')

    if [[ -z "${VERSION}" ]]; then
        echo "failed to get latest version from GitHub API"
        exit 1
    fi

    install_gale "${VERSION}" "${os}" "${arch}"
}

main "$@"