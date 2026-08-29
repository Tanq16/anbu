package utils

import (
	"os"

	"charm.land/lipgloss/v2"
	"github.com/rs/zerolog/log"
)

func PrintSuccess(text string) {
	if GlobalDebugFlag {
		log.Info().Msg(text)
		return
	}
	lipgloss.Println(successStyle.Render(StyleSymbols["pass"] + " " + text))
}

func PrintError(text string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Error().Err(err).Msg(text)
		} else {
			log.Error().Msg(text)
		}
		return
	}
	lipgloss.Println(errorStyle.Render(StyleSymbols["fail"] + " " + text))
}

func PrintFatal(text string, err error) {
	PrintError(text, err)
	os.Exit(1)
}

func PrintWarn(text string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Warn().Err(err).Msg(text)
		} else {
			log.Warn().Msg(text)
		}
		return
	}
	lipgloss.Println(warningStyle.Render(StyleSymbols["warning"] + " " + text))
}

func PrintInfo(text string) {
	if GlobalDebugFlag {
		log.Info().Msg(text)
		return
	}
	lipgloss.Println(infoStyle.Render(StyleSymbols["arrow"] + " " + text))
}

func PrintDebug(text string) {
	if GlobalDebugFlag {
		log.Debug().Msg(text)
		return
	}
	lipgloss.Println(debugStyle.Render(text))
}

func PrintStream(text string) {
	if GlobalDebugFlag {
		log.Debug().Msg(text)
		return
	}
	lipgloss.Println(streamStyle.Render(text))
}

func PrintGeneric(text string) {
	lipgloss.Println(text)
}

func FSuccess(text string) string {
	if GlobalDebugFlag {
		return text
	}
	return successStyle.Render(text)
}

func FError(text string) string {
	if GlobalDebugFlag {
		return text
	}
	return errorStyle.Render(text)
}

func FWarning(text string) string {
	if GlobalDebugFlag {
		return text
	}
	return warningStyle.Render(text)
}

func FInfo(text string) string {
	if GlobalDebugFlag {
		return text
	}
	return infoStyle.Render(text)
}

func FDebug(text string) string {
	if GlobalDebugFlag {
		return text
	}
	return debugStyle.Render(text)
}

func FStream(text string) string {
	if GlobalDebugFlag {
		return text
	}
	return streamStyle.Render(text)
}

func FGeneric(text string) string {
	return text
}

func LineBreak() {
	lipgloss.Println()
}
