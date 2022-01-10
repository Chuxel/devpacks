FROM docker.io/paketobuildpacks/run:full-cnb

USER root
RUN bash -c "$(curl -fsSL "https://raw.githubusercontent.com/microsoft/vscode-dev-containers/main/script-library/common-debian.sh")" -- true cnb 1000 1000 false true true \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*
USER cnb
