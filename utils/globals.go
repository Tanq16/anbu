package utils

import (
	"os"

	"github.com/charmbracelet/x/term"
)

var GlobalDebugFlag bool

var (
	StdinIsTerminal  = term.IsTerminal(os.Stdin.Fd())
	StdoutIsTerminal = term.IsTerminal(os.Stdout.Fd())
)
