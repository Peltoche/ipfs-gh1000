package index

import (
	"context"
	"fmt"

	"github.com/Peltoche/ipfs-gh1000/pkg/ipfs"
	"github.com/Peltoche/ipfs-gh1000/pkg/metadata"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/teris-io/cli"
)

func PurgeCmd() cli.Command {
	return cli.NewCommand("purge",
		"Remove all the entries from the index").
		WithAction(purgeAction)
}

func purgeAction(args []string, options map[string]string) int {
	ctx := context.Background()

	shell := shell.NewLocalShell()

	ipfsIndexer, err := ipfs.NewIndexer(shell, "gh1000")
	if err != nil {
		fmt.Println(err)
		return 1
	}

	err = ipfsIndexer.SaveIndex(ctx, map[string]metadata.RepoMetadata{})
	if err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}
