package index

import (
	"context"
	"fmt"
	"os"

	"github.com/Peltoche/ipfs-gh1000/pkg/ipfs"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/teris-io/cli"
)

func CatCmd() cli.Command {
	return cli.NewCommand("cat",
		"Print the content index content").
		WithAction(catAction)
}

func catAction(args []string, options map[string]string) int {
	ctx := context.Background()

	shell := shell.NewLocalShell()

	ipfsIndexer, err := ipfs.NewIndexer(shell, "gh1000")
	if err != nil {
		fmt.Println(err)
		return 1
	}

	index, err := ipfsIndexer.RetrieveIndex(ctx)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	err = ipfsIndexer.EncodeIndex(index, os.Stdout)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	fmt.Println("")

	return 0
}
