package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/hevinxx/private-skill-repository/internal/skillrc"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show installed skills and update availability",
		RunE:  runStatus,
	}
	cmd.Flags().Bool("global", false, "Show global installations")
	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	global, _ := cmd.Flags().GetBool("global")
	scope := "project"
	if global {
		scope = "global"
	}

	rcPath, err := skillrc.Path(scope, buildInfo.BinaryName)
	if err != nil {
		return err
	}
	rc, err := skillrc.Load(rcPath)
	if err != nil {
		return err
	}

	if len(rc.Installed) == 0 {
		fmt.Println("No skills installed.")
		return nil
	}

	client, err := newGitHubClient()
	if err != nil {
		return err
	}

	latestCommit, err := client.GetLatestCommit()
	if err != nil {
		return err
	}

	idx, err := fetchIndex()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tSTATUS")

	for _, s := range rc.Installed {
		status := "up to date"
		if s.Commit != latestCommit {
			path := entryPath(idx, s.Name)
			changed, err := client.HasFileChanged(path, s.Commit, latestCommit)
			if err != nil {
				status = "error checking"
			} else if changed {
				status = "update available"
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.Type, status)
	}
	w.Flush()
	return nil
}
