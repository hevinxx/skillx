package cmd

import (
	"fmt"

	"github.com/hevinxx/private-skill-repository/internal/installer"
	"github.com/hevinxx/private-skill-repository/internal/skillrc"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed skill",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}
	cmd.Flags().Bool("global", false, "Remove from global installation")
	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	global, _ := cmd.Flags().GetBool("global")

	scope := "project"
	if global {
		scope = "global"
	}

	// Find in .skillrc
	rcPath, err := skillrc.Path(scope, buildInfo.BinaryName)
	if err != nil {
		return err
	}
	rc, err := skillrc.Load(rcPath)
	if err != nil {
		return err
	}
	existing := rc.Find(name)
	if existing == nil {
		return fmt.Errorf("skill '%s' is not installed", name)
	}

	// Remove files
	client, err := newGitHubClient()
	if err != nil {
		return err
	}
	inst := installer.New(client, buildInfo.BinaryName)
	if err := inst.Remove(name, existing.Type, scope); err != nil {
		return fmt.Errorf("removing '%s': %w", name, err)
	}

	// Update tracking
	if err := installer.TrackRemove(name, scope, buildInfo.BinaryName); err != nil {
		return fmt.Errorf("updating tracking: %w", err)
	}

	fmt.Printf("Removed '%s'.\n", name)
	return nil
}
