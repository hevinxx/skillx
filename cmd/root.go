package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// BuildInfo holds build-time configuration values.
type BuildInfo struct {
	BinaryName  string
	DefaultOrg  string
	DefaultRepo string
	DefaultHost string
	Version     string
}

var buildInfo BuildInfo

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   buildInfo.BinaryName,
		Short: "Manage Claude Code skills from your organization's skill repository",
		Long: fmt.Sprintf(`%s is a CLI tool that connects Claude Code developers to their
organization's shared skill repository. Install, manage, and contribute
skills with simple commands.`, buildInfo.BinaryName),
		SilenceUsage: true,
	}

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newInitRepoCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newInfoCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newRemoveCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newCreateCmd())

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", buildInfo.BinaryName, buildInfo.Version)
		},
	}
}

// Execute runs the root command.
func Execute(info BuildInfo) {
	buildInfo = info
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
