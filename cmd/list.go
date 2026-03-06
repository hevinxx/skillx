package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/hevinxx/skillx/internal/registry"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available skills",
		RunE:  runList,
	}
	cmd.Flags().String("type", "", "Filter by type (command, skill, agent)")
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	idx, err := fetchIndex()
	if err != nil {
		return err
	}

	skills := idx.Skills
	if t, _ := cmd.Flags().GetString("type"); t != "" {
		skills = idx.FilterByType(t)
	}

	if len(skills) == 0 {
		fmt.Println("No skills found.")
		return nil
	}

	printSkillTable(skills)
	return nil
}

func printSkillTable(skills []registry.SkillEntry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tDESCRIPTION")
	for _, s := range skills {
		desc := s.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.Type, desc)
	}
	w.Flush()
}
