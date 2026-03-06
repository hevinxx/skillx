package cmd

import (
	"fmt"

	"github.com/hevinxx/skillx/internal/template"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Scaffold a new skill from template",
		Long:  "Creates a new skill directory with skill.yaml and a placeholder file in the local skill repo clone.",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
	}
	cmd.Flags().String("type", "", "Skill type: command, skill, or agent (required)")
	cmd.MarkFlagRequired("type")
	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	skillType, _ := cmd.Flags().GetString("type")

	if skillType != "command" && skillType != "skill" && skillType != "agent" {
		return fmt.Errorf("invalid type '%s': must be command, skill, or agent", skillType)
	}

	created, err := template.ScaffoldSkill(".", name, skillType)
	if err != nil {
		return err
	}

	fmt.Printf("Created '%s' (%s):\n", name, skillType)
	for _, f := range created {
		fmt.Printf("  %s\n", f)
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %ss/%s/skill.yaml — fill in description, author, tags\n", skillType, name)
	fmt.Printf("  2. Edit %ss/%s/%s.md — write the skill content\n", skillType, name, name)
	fmt.Println("  3. Commit and open a pull request")
	return nil
}
