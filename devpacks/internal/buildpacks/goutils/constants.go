package goutils

const BUILDPACK_NAME = "goutils"

// Go tools that are isImportant && !replacedByGopls based on https://github.com/golang/vscode-go/blob/v0.31.1/src/goToolsInformation.ts
const DEFAULT_GO_UTILS = "golang.org/x/tools/gopls@latest honnef.co/go/tools/cmd/staticcheck@latest golang.org/x/lint/golint@latest github.com/mgechev/revive@latest github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest github.com/ramya-rao-a/go-outline@latest github.com/go-delve/delve/cmd/dlv@latest"
