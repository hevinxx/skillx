package cmd

import (
	"fmt"

	"github.com/hevinxx/skillx/internal/installer"
"github.com/hevinxx/skillx/internal/skillrc"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update installed skills to latest version",
		Long:  "Update all installed skills, or a specific one if named.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runUpdate,
	}
	cmd.Flags().Bool("global", false, "Update global installations")
	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
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

	client, err := newProvider()
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

	inst := installer.New(client, buildInfo.BinaryName)
	var updated int

	targets := rc.Installed
	if len(args) > 0 {
		s := rc.Find(args[0])
		if s == nil {
			return fmt.Errorf("skill '%s' is not installed", args[0])
		}
		targets = []skillrc.InstalledSkill{*s}
	}

	for _, s := range targets {
		if s.Commit == latestCommit {
			continue
		}

		changed, err := client.HasFileChanged(entryPath(idx, s.Name), s.Commit, latestCommit)
		if err != nil {
			fmt.Printf("  %s: error checking changes: %v\n", s.Name, err)
			continue
		}
		if !changed {
			continue
		}

		entry := idx.Find(s.Name)
		if entry == nil {
			fmt.Printf("  %s: no longer in index, skipping\n", s.Name)
			continue
		}

		if _, err := inst.Install(entry, scope); err != nil {
			fmt.Printf("  %s: error updating: %v\n", s.Name, err)
			continue
		}

		if err := installer.TrackInstall(s.Name, s.Type, latestCommit, scope, buildInfo.BinaryName); err != nil {
			fmt.Printf("  %s: error tracking: %v\n", s.Name, err)
			continue
		}

		fmt.Printf("  Updated: %s\n", s.Name)
		updated++
	}

	if updated == 0 {
		fmt.Println("All skills are up to date.")
	} else {
		fmt.Printf("%d skill(s) updated.\n", updated)
	}
	return nil
}
