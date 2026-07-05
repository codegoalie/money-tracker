package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "moneytracker",
		Usage: "cash flow forecast from LunchMoney data",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Phase 0: no behaviors exist yet. Exit 2 (operational error)
			// for every invocation; real commands land in later phases.
			return cli.Exit(fmt.Sprintf("moneytracker: unknown command %q", cmd.Args().First()), 2)
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
