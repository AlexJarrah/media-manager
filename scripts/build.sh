#!/bin/bash

if [ "$(uname)" != "Linux" ]; then
  echo "Unsupported operating system."
  exit
fi

# Required packages for various package managers
APT_PKGS="git golang"
DNF_PKGS="git golang"
PACMAN_PKGS="git go"

# Detect the system's package manager and install/update necessary packages
if command -v apt &>/dev/null; then
    echo "Installing dependencies: $APT_PKGS"
    sudo apt install $APT_PKGS -y

elif command -v dnf &>/dev/null; then
    echo "Installing dependencies: $DNF_PKGS"
    sudo dnf install $DNF_PKGS -y

elif command -v pacman &>/dev/null; then
    echo "Installing dependencies: $PACMAN_PKGS"
    sudo pacman -S $PACMAN_PKGS --noconfirm

else
    echo "Unsupported package manager"
    exit
fi

# Clone the repository into the current directory
git clone https://gitlab.com/AlexJarrah/media-manager.git

# Navigate into the repository
cd media-manager

# Install required Go packages
go mod tidy

# Build binary
go build -o media-manager cmd/main.go
