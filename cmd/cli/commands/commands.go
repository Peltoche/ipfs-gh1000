package commands

import (
	"github.com/Peltoche/ipfs-gh1000/cmd/cli/commands/index"
	"github.com/teris-io/cli"
)

func NewApp() cli.App {

	return cli.New("gh1000-cli").
		WithCommand(index.IndexCmd())
}
