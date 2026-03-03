package utils

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

func PrintSuccess(text string) {
	if GlobalDebugFlag {
		log.Info().Str("package", "utils").Msg(text)
	} else if GlobalForAIFlag {
		fmt.Println("[OK] " + text)
	} else {
		fmt.Println(successStyle.Render(StyleSymbols["pass"] + " " + text))
	}
}
func PrintError(text string, err error) {
	if GlobalDebugFlag {
		log.Error().Str("package", "utils").Err(err).Msg(text)
	} else if GlobalForAIFlag {
		fmt.Println("[ERROR] " + text)
	} else {
		fmt.Println(errorStyle.Render(StyleSymbols["fail"] + " " + text))
	}
}
func PrintFatal(text string, err error) {
	if GlobalDebugFlag {
		log.Fatal().Str("package", "utils").Err(err).Msg(text)
	} else if GlobalForAIFlag {
		fmt.Println("[FATAL] " + text)
		os.Exit(1)
	} else {
		fmt.Println(errorStyle.Render(StyleSymbols["fail"] + " " + text))
		os.Exit(1)
	}
}
func PrintWarn(text string, err error) {
	if GlobalDebugFlag {
		log.Warn().Str("package", "utils").Err(err).Msg(text)
	} else if GlobalForAIFlag {
		fmt.Println("[WARN] " + text)
	} else {
		fmt.Println(warningStyle.Render(StyleSymbols["warning"] + " " + text))
	}
}
func PrintInfo(text string) {
	if GlobalDebugFlag {
		log.Info().Str("package", "utils").Msg(text)
	} else if GlobalForAIFlag {
		fmt.Println("[INFO] " + text)
	} else {
		fmt.Println(infoStyle.Render(StyleSymbols["arrow"] + " " + text))
	}
}
func PrintDebug(text string) {
	if GlobalDebugFlag {
		log.Debug().Str("package", "utils").Msg(text)
	} else if GlobalForAIFlag {
		fmt.Println("[DEBUG] " + text)
	} else {
		fmt.Println(debugStyle.Render(text))
	}
}
func PrintStream(text string) {
	if GlobalDebugFlag {
		log.Debug().Str("package", "utils").Msg(text)
	} else if GlobalForAIFlag {
		fmt.Println(text)
	} else {
		fmt.Println(streamStyle.Render(text))
	}
}
func PrintGeneric(text string) {
	if GlobalDebugFlag {
		log.Debug().Str("package", "utils").Msg(text)
	} else {
		fmt.Println(text)
	}
}
func FSuccess(text string) string {
	if GlobalDebugFlag || GlobalForAIFlag {
		return text
	}
	return successStyle.Render(text)
}
func FError(text string) string {
	if GlobalDebugFlag || GlobalForAIFlag {
		return text
	}
	return errorStyle.Render(text)
}
func FWarning(text string) string {
	if GlobalDebugFlag || GlobalForAIFlag {
		return text
	}
	return warningStyle.Render(text)
}
func FInfo(text string) string {
	if GlobalDebugFlag || GlobalForAIFlag {
		return text
	}
	return infoStyle.Render(text)
}
func FDebug(text string) string {
	if GlobalDebugFlag || GlobalForAIFlag {
		return text
	}
	return debugStyle.Render(text)
}
func FStream(text string) string {
	if GlobalDebugFlag || GlobalForAIFlag {
		return text
	}
	return streamStyle.Render(text)
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
