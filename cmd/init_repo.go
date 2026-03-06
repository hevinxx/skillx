package cmd

import (
	"fmt"

	"github.com/hevinxx/private-skill-repository/internal/template"
	"github.com/spf13/cobra"
)

func newInitRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-repo [directory]",
		Short: "Create a new skill repository from template",
		Long:  "Scaffolds a new skill repository with the standard directory structure, CI workflows, and documentation.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInitRepo,
	}
	return cmd
}

func runInitRepo(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	created, err := template.InitRepo(dir)
	if err != nil {
		return err
	}

	fmt.Println("Skill repository initialized:")
	for _, f := range created {
		fmt.Printf("  %s\n", f)
	}
	fmt.Println("  commands/")
	fmt.Println("  skills/")
	fmt.Println("  agents/")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Create a private GitHub repository")
	fmt.Println("  2. Push this directory to the repository")
	fmt.Printf("  3. Run '%s init' on each developer's machine\n", buildInfo.BinaryName)
	return nil
}
