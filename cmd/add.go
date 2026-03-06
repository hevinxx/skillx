package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hevinxx/private-skill-repository/internal/installer"
	"github.com/hevinxx/private-skill-repository/internal/skillrc"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Install a skill",
		Args:  cobra.ExactArgs(1),
		RunE:  runAdd,
	}
	cmd.Flags().Bool("global", false, "Install globally (~/.claude/) instead of project-local")
	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	global, _ := cmd.Flags().GetBool("global")

	scope := "project"
	if global {
		scope = "global"
	}

	// Check if already installed
	rcPath, err := skillrc.Path(scope, buildInfo.BinaryName)
	if err != nil {
		return err
	}
	rc, err := skillrc.Load(rcPath)
	if err != nil {
		return err
	}
	if existing := rc.Find(name); existing != nil {
		fmt.Printf("'%s' is already installed. Overwrite? [y/N]: ", name)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Fetch index and find skill
	idx, err := fetchIndex()
	if err != nil {
		return err
	}
	entry := idx.Find(name)
	if entry == nil {
		return fmt.Errorf("skill '%s' not found", name)
	}

	// Install files
	client, err := newGitHubClient()
	if err != nil {
		return err
	}
	inst := installer.New(client, buildInfo.BinaryName)
	files, err := inst.Install(entry, scope)
	if err != nil {
		return fmt.Errorf("installing '%s': %w", name, err)
	}

	// Get commit hash and track
	commit, err := client.GetLatestCommit()
	if err != nil {
		return fmt.Errorf("getting commit hash: %w", err)
	}
	if err := installer.TrackInstall(name, entry.Type, commit, scope, buildInfo.BinaryName); err != nil {
		return fmt.Errorf("tracking installation: %w", err)
	}

	fmt.Printf("Installed '%s' (%s):\n", name, entry.Type)
	for _, f := range files {
		fmt.Printf("  %s\n", f)
	}
	return nil
}
