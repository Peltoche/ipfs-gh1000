package index

import (
	"github.com/teris-io/cli"
)

func IndexCmd() cli.Command {
	return cli.NewCommand("index",
		"Manipulate the index").
		WithCommand(PurgeCmd())
}
