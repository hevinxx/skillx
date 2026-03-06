package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show detailed information about a skill",
		Args:  cobra.ExactArgs(1),
		RunE:  runInfo,
	}
}

func runInfo(cmd *cobra.Command, args []string) error {
	idx, err := fetchIndex()
	if err != nil {
		return err
	}

	entry := idx.Find(args[0])
	if entry == nil {
		return fmt.Errorf("skill '%s' not found", args[0])
	}

	fmt.Printf("Name:        %s\n", entry.Name)
	fmt.Printf("Type:        %s\n", entry.Type)
	fmt.Printf("Description: %s\n", entry.Description)
	fmt.Printf("Author:      %s\n", entry.Author)
	fmt.Printf("Path:        %s\n", entry.Path)
	if len(entry.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(entry.Tags, ", "))
	}
	return nil
}
