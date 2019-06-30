package metadata

import (
	"fmt"
	"runtime"
)


// Variables defined by the Makefile and passed in with ldflags
var Version string

func GetVersion() string {
	if Version == "" {
		Version = "dev-build"
	}
	return fmt.Sprintf("Version: %s\n Go version: %s\n OS/Arch: %s/%s",
		Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}