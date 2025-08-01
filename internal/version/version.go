package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time using ldflags
var (
	// Version is the semantic version of bc4
	Version = "dev"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// BuildDate is the date when the binary was built
	BuildDate = "unknown"

	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)

// Info represents version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("bc4 version %s (%s, %s)", i.Version, i.GitCommit, i.BuildDate)
}

// DetailedString returns a detailed version string
func (i Info) DetailedString() string {
	return fmt.Sprintf(`bc4 version %s
  Commit:     %s
  Built:      %s
  Go version: %s
  Platform:   %s`,
		i.Version,
		i.GitCommit,
		i.BuildDate,
		i.GoVersion,
		i.Platform,
	)
}

// UserAgent returns a properly formatted User-Agent string for HTTP requests
func UserAgent() string {
	return fmt.Sprintf("bc4-cli/%s (github.com/needmore/bc4)", Version)
}
