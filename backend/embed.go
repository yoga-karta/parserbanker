package main

import "embed"

// webui berisi hasil `next build` (static export) dari frontend, di-copy ke
// folder ini oleh CI (lihat .github/workflows/build-windows.yml) sebelum
// `go build`. Folder ini sengaja tidak dicommit ke git (lihat .gitignore) -
// isinya generated, bukan source.
//
//go:embed all:webui
var webUI embed.FS
