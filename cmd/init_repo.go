package cmd

import (
	"fmt"

	"github.com/hevinxx/skillx/internal/template"
	"github.com/spf13/cobra"
)

func newInitRepoCmd() *cobra.Command {
	var providerType string

	cmd := &cobra.Command{
		Use:   "init-repo [directory]",
		Short: "Create a new skill repository from template",
		Long:  "Scaffolds a new skill repository with the standard directory structure, CI workflows, and documentation.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			switch providerType {
			case "github", "gitlab", "gitea":
			default:
				return fmt.Errorf("unsupported provider type %q: must be github, gitlab, or gitea", providerType)
			}

			created, err := template.InitRepo(dir, providerType)
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
			fmt.Printf("  1. Create a private %s repository\n", providerType)
			fmt.Println("  2. Push this directory to the repository")
			fmt.Printf("  3. Run '%s init' on each developer's machine\n", buildInfo.BinaryName)

			if providerType == "gitlab" {
				fmt.Println()
				fmt.Println("Note: Set CI_PUSH_TOKEN variable in GitLab CI/CD settings")
				fmt.Println("  to allow the pipeline to push index.yaml commits.")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&providerType, "provider", "p", "github", "Git hosting provider (github, gitlab, gitea)")

	return cmd
}
