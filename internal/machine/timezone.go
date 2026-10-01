package machine

import (
	"os"
	"strings"
)

const (
	defaultTimezone = "Etc/UTC"
	localtimePath   = "/etc/localtime"
)

func LocalTimezone() string {
	// time.Local renames whatever zone it loaded to "Local", so the IANA name comes from the symlink instead.
	link, err := os.Readlink(localtimePath)
	if err != nil {
		return defaultTimezone
	}
	if _, zone, found := strings.Cut(link, "/zoneinfo/"); found && zone != "" {
		return zone
	}
	return defaultTimezone
}
