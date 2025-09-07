package version

import (
	"fmt"
	"runtime"
)

// Version is set by build flags
var Version = "dev"

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("artship version %s %s/%s (built with %s)",
		i.Version, i.Platform, i.Arch, i.GoVersion)
}
