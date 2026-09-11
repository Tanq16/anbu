package archive

import "regexp"

type CreateConfig struct {
	Paths    []string
	Output   string
	Include  []*regexp.Regexp
	Exclude  []*regexp.Regexp
	Bare     bool
	Encrypt  bool
	Password string
}

type ExtractConfig struct {
	Archive  string
	Dest     string
	Bare     bool
	Password string
}
