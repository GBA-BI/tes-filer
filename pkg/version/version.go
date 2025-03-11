package version

import (
	"fmt"
	"os"
	"runtime"

	"github.com/gosuri/uitable"
)

var (
	// Module ...
	Module string
	// Version is the code version.
	Version string
	// GitBranch code branch.
	GitBranch string
	// GitCommit is the git commit.
	GitCommit string
	// GitTreeState clean/dirty.
	GitTreeState string
	// BuildTime is the build time.
	BuildTime string
)

type Info struct {
	Module       string `json:"module"`
	Version      string `json:"version"`
	GitBranch    string `json:"gitBranch"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildTime    string `json:"buildTime"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// String return the string of Info.
func (info Info) String() string {
	table := uitable.New()
	table.RightAlign(0)
	table.MaxColWidth = 80
	table.Separator = " "
	table.AddRow("module:", info.Module)
	table.AddRow("version:", info.Version)
	table.AddRow("gitCommit:", info.GitCommit)
	table.AddRow("gitBranch:", info.GitBranch)
	table.AddRow("gitTreeState:", info.GitTreeState)
	table.AddRow("buildTime:", info.BuildTime)
	table.AddRow("goVersion:", info.GoVersion)
	table.AddRow("compiler:", info.Compiler)
	table.AddRow("platform:", info.Platform)

	return table.String()
}

func Get() Info {
	return Info{
		Module:       Module,
		Version:      Version,
		GitBranch:    GitBranch,
		GitCommit:    GitCommit,
		GitTreeState: GitTreeState,
		BuildTime:    BuildTime,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// PrintVersionOrContinue will print git commit and exit with os.Exit(0) if CLI v flag is present.
func PrintVersionOrContinue() {
	fmt.Printf("%s\n", Get())
	if versionFlag {
		os.Exit(0)
	}
}
