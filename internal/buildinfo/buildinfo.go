package buildinfo

import "strings"

var (
	Version        = "dev"
	InstallMethod  = "source"
	UpdatesEnabled = "true"
)

func UpdateChecksEnabled() bool {
	return !strings.EqualFold(UpdatesEnabled, "false") && Version != "dev"
}
