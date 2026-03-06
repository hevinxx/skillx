# SkillX

A CLI tool for managing shared Claude Code skills across your organization.

Organizations maintain a private GitHub repository of skills (commands, behavioral instructions, agent definitions). SkillX provides the tooling to install, manage, and contribute skills from that repository.

## Install

### From source

```bash
go install github.com/hevinxx/private-skill-repository@latest
```

### From releases

Download the binary for your platform from [Releases](https://github.com/hevinxx/private-skill-repository/releases).

### Custom binary name

Build with a custom name for your organization:

```bash
go build -ldflags "-X main.binaryName=mycompany-skills -X main.defaultOrg=mycompany -X main.defaultRepo=our-skills" -o mycompany-skills .
```

## Quick Start

```bash
# 1. Configure the CLI (points to your org's skill repository)
skillx init

# 2. Browse available skills
skillx list
skillx search "review"

# 3. Install a skill into your project
skillx add code-review

# 4. Check for updates
skillx status
skillx update
```

## Setting Up Your Organization's Skill Repository

```bash
# Create the skill repo structure
mkdir my-org-skills && cd my-org-skills
skillx init-repo

# Push to GitHub as a private repo
git init && git add . && git commit -m "Initial skill repository"
gh repo create my-org/claude-skills --private --source=. --push
```

## Contributing Skills

```bash
# In your local clone of the skill repository
skillx create my-new-skill --type command

# Edit the generated files, then open a PR
```

## Skill Types

| Type | Install Path | Description |
|------|-------------|-------------|
| `command` | `.claude/commands/` | Slash commands for Claude Code |
| `skill` | `.claude/skills/` | Behavioral instructions and domain knowledge |
| `agent` | `.claude/agents/` | Autonomous agent definitions |

## Commands

| Command | Description |
|---------|-------------|
| `skillx init` | Configure CLI (GitHub org, repo, auth) |
| `skillx init-repo` | Scaffold a new skill repository |
| `skillx list [--type TYPE]` | List available skills |
| `skillx search QUERY` | Search by name, description, or tags |
| `skillx info NAME` | Show skill details |
| `skillx add NAME [--global]` | Install a skill |
| `skillx remove NAME [--global]` | Remove a skill |
| `skillx update [NAME] [--global]` | Update to latest |
| `skillx status [--global]` | Check update availability |
| `skillx create NAME --type TYPE` | Scaffold a new skill |

## Authentication

SkillX needs access to your private skill repository. It checks for a token in this order:

1. `SKILLX_GITHUB_TOKEN` environment variable
2. `GITHUB_TOKEN` environment variable
3. GitHub CLI (`gh auth token`)

## License

MIT
