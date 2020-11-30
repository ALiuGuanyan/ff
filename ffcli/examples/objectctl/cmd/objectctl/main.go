package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ALiuGuanyan/ff/v3/ffcli"
	"github.com/ALiuGuanyan/ff/v3/ffcli/examples/objectctl/pkg/createcmd"
	"github.com/ALiuGuanyan/ff/v3/ffcli/examples/objectctl/pkg/deletecmd"
	"github.com/ALiuGuanyan/ff/v3/ffcli/examples/objectctl/pkg/listcmd"
	"github.com/ALiuGuanyan/ff/v3/ffcli/examples/objectctl/pkg/objectapi"
	"github.com/ALiuGuanyan/ff/v3/ffcli/examples/objectctl/pkg/rootcmd"
)

func main() {
	var (
		out                     = os.Stdout
		rootCommand, rootConfig = rootcmd.New()
		createCommand           = createcmd.New(rootConfig, out)
		deleteCommand           = deletecmd.New(rootConfig, out)
		listCommand             = listcmd.New(rootConfig, out)
	)

	rootCommand.Subcommands = []*ffcli.Command{
		createCommand,
		deleteCommand,
		listCommand,
	}
	flag.Parse()
	if err := rootCommand.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error during Parse: %v\n", err)
		os.Exit(1)
	}

	client, err := objectapi.NewClient(rootConfig.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error constructing object API client: %v\n", err)
		os.Exit(1)
	}

	rootConfig.Client = client

	if err := rootCommand.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
