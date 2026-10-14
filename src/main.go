package main

import (
	"embed"
	"github.com/multihubandredye-prog/W-C-A_GO/cmd"
)

//go:embed views/index.html
var embedIndex embed.FS

//go:embed views
var embedViews embed.FS

func main() {
	// Trigger build v3
	cmd.Execute(embedIndex, embedViews)
}
