// Command penbridge is a CLI bridge to the SiYuan kernel API. It lets you (and
// AI agents) operate every open API endpoint in a safe, scriptable way.
//
// Install:
//
//	go install github.com/zerx-lab/penbridge-cli/cmd/penbridge@latest
package main

import "github.com/zerx-lab/penbridge-cli/internal/cli"

func main() {
	cli.Execute()
}
