package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Display detailed version and build information.

Shows:
  • Version number
  • Git commit
  • Build date
  • Go version
  • Platform/OS

Examples:
  codex-router version
  codex-router version --output json`,
	Run: func(cmd *cobra.Command, args []string) {
		info := VersionInfo{
			Version:   Version,
			Commit:    Commit,
			BuildDate: Date,
			GoVersion: runtime.Version(),
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		}

		if globalOpts.Output == "json" {
			data, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("codex-router %s\n", info.Version)
			fmt.Printf("  Commit:     %s\n", info.Commit)
			fmt.Printf("  Built:      %s\n", info.BuildDate)
			fmt.Printf("  Go version: %s\n", info.GoVersion)
			fmt.Printf("  Platform:   %s\n", info.Platform)
		}
	},
}

// VersionInfo holds version metadata
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
