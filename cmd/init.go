package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hevinxx/skillx/internal/config"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Set up CLI configuration",
		Long:  "Interactive setup that creates the configuration file with GitHub org, repo, and authentication settings.",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Setting up " + buildInfo.BinaryName + " configuration")
	fmt.Println()

	host := prompt(reader, "GitHub host", defaultVal(buildInfo.DefaultHost, "github.com"))
	org := prompt(reader, "GitHub organization", buildInfo.DefaultOrg)
	repo := prompt(reader, "Skill repository name", defaultVal(buildInfo.DefaultRepo, "claude-skills"))
	scope := prompt(reader, "Default install scope (project/global)", "project")

	if scope != "project" && scope != "global" {
		scope = "project"
	}

	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Host: host,
			Org:  org,
			Repo: repo,
		},
		Defaults: config.Defaults{
			Scope: scope,
		},
	}

	if err := config.Save(buildInfo.BinaryName, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	p, _ := config.Path(buildInfo.BinaryName)
	fmt.Printf("\nConfiguration saved to %s\n", p)
	return nil
}

func prompt(reader *bufio.Reader, label, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("%s: ", label)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func defaultVal(buildTime, fallback string) string {
	if buildTime != "" {
		return buildTime
	}
	return fallback
}
