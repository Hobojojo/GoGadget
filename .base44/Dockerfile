FROM golang:1.23-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends entr curl \
    && rm -rf /var/lib/apt/lists/* \
    && curl -fL --retry 3 https://github.com/tsl0922/ttyd/releases/download/1.7.7/ttyd.x86_64 -o /usr/local/bin/ttyd \
    && echo '8a217c968aba172e0dbf3f34447218dc015bc4d5e59bf51db2f2cd12b7be4f55  /usr/local/bin/ttyd' | sha256sum -c - \
    && chmod 755 /usr/local/bin/ttyd \
    && useradd --create-home --uid 1000 preview \
    && mkdir -p /home/preview/go /home/preview/.cache/go-build /home/preview/.config \
    && chown -R preview:preview /home/preview

ENV HOME=/home/preview \
    GOPATH=/home/preview/go \
    GOCACHE=/home/preview/.cache/go-build
WORKDIR /app
# Build/watch as root because the imported checkout is root-owned (mode 0700).
# ttyd drops to the preview user before accepting terminal sessions.
