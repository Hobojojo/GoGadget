#!/bin/sh
set -eu

go mod download

# Watch the mounted checkout, rebuilding and reconnecting terminal sessions on edits.
# entr exits with status 2 when a new file requires refreshing the watch list.
while :; do
    status=0
    find . -type f \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \) \
        | entr -d -n -r sh -ec '
            go build -mod=readonly -o /tmp/tui-launcher .
            echo "Serving live source through the browser terminal on port 3000"
            cd /home/preview
            exec ttyd -W -u 1000 -g 1000 -i 0.0.0.0 -p 3000 -t rendererType=dom -t "titleFixed=TUI App Launcher" /tmp/tui-launcher
        ' || status=$?
    [ "$status" -eq 2 ] || exit "$status"
done
