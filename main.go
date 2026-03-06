package main

import "github.com/hevinxx/private-skill-repository/cmd"

// Build-time variables for customization
var (
	binaryName  = "skillx"
	defaultOrg  = ""
	defaultRepo = ""
	defaultHost = "github.com"
	version     = "dev"
)

func main() {
	cmd.Execute(cmd.BuildInfo{
		BinaryName:  binaryName,
		DefaultOrg:  defaultOrg,
		DefaultRepo: defaultRepo,
		DefaultHost: defaultHost,
		Version:     version,
	})
}
