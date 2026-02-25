package utils

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

func PrintSuccess(text string) {
	if !GlobalDebugFlag {
		fmt.Println(successStyle.Render(text))
	} else {
		log.Info().Msg(text)
	}
}
func PrintError(text string, err error) {
	if !GlobalDebugFlag {
		fmt.Println(errorStyle.Render(text))
	} else {
		log.Error().Err(err).Msg(text)
	}
}
func PrintFatal(text string, err error) {
	if !GlobalDebugFlag {
		fmt.Println(errorStyle.Render(text))
		os.Exit(1)
	} else {
		log.Fatal().Err(err).Msg(text)
	}
}
func PrintWarn(text string, err error) {
	if !GlobalDebugFlag {
		fmt.Println(warningStyle.Render(text))
	} else {
		log.Warn().Err(err).Msg(text)
	}
}
func PrintInfo(text string) {
	if !GlobalDebugFlag {
		fmt.Println(infoStyle.Render(text))
	} else {
		log.Info().Msg(text)
	}
}
func PrintDebug(text string) {
	if !GlobalDebugFlag {
		fmt.Println(debugStyle.Render(text))
	} else {
		log.Debug().Msg(text)
	}
}
func PrintStream(text string) {
	if !GlobalDebugFlag {
		fmt.Println(streamStyle.Render(text))
	} else {
		log.Debug().Msg(text)
	}
}
func PrintGeneric(text string) {
	if !GlobalDebugFlag {
		fmt.Println(text)
	} else {
		log.Debug().Msg(text)
	}
}
func FSuccess(text string) string {
	if !GlobalDebugFlag {
		return successStyle.Render(text)
	}
	return text
}
func FError(text string) string {
	if !GlobalDebugFlag {
		return errorStyle.Render(text)
	}
	return text
}
func FWarning(text string) string {
	if !GlobalDebugFlag {
		return warningStyle.Render(text)
	}
	return text
}
func FInfo(text string) string {
	if !GlobalDebugFlag {
		return infoStyle.Render(text)
	}
	return text
}
func FDebug(text string) string {
	if !GlobalDebugFlag {
		return debugStyle.Render(text)
	}
	return text
}
func FStream(text string) string {
	if !GlobalDebugFlag {
		return streamStyle.Render(text)
	}
	return text
}
func FGeneric(text string) string {
	return text
}

func LineBreak() {
	fmt.Println()
}
func ClearTerminal(lines int) {
	if lines > 0 {
		fmt.Printf("\033[%dA\r\033[K", lines)
	}
}
