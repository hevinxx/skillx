package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search skills by name, description, or tags",
		Args:  cobra.ExactArgs(1),
		RunE:  runSearch,
	}
}

func runSearch(cmd *cobra.Command, args []string) error {
	idx, err := fetchIndex()
	if err != nil {
		return err
	}

	results := idx.Search(args[0])
	if len(results) == 0 {
		fmt.Printf("No skills matching '%s'.\n", args[0])
		return nil
	}

	printSkillTable(results)
	return nil
}
