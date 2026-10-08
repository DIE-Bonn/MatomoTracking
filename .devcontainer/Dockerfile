# https://mcr.microsoft.com/v2/vscode/devcontainers/go/tags/list
FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

# make, curl, and Yarn are preinstalled; Yarn's expired repository key breaks apt in features.
RUN rm -f /etc/apt/sources.list.d/yarn.list
