package anbuGenerics

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tanq16/anbu/utils"
)

func ProcessRange(input string) ([]int, error) {
	logger := utils.GetLogger("loopcmd")
	var rangeElems []int
	if strings.Contains(input, "-") {
		logger.Debug().Str("input", input).Msg("processing range")
		parts := strings.Split(input, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format: %s", input)
		}
		start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid start of range: %w", err)
		}
		end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid end of range: %w", err)
		}
		if start > end {
			return nil, fmt.Errorf("start of range cannot be greater than end: %s", input)
		}
		loopRange := make([]int, end-start+1)
		for i := start; i <= end; i++ {
			loopRange[i-start] = i
		}
		rangeElems = loopRange
		logger.Debug().Str("range", fmt.Sprintf("%d-%d", start, end)).Msg("range processed")
	} else {
		logger.Debug().Str("input", input).Msg("processing count")
		count, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			return nil, fmt.Errorf("invalid count: %w", err)
		}
		if count < 0 {
			return nil, fmt.Errorf("count must be non-negative")
		}
		loopRange := make([]int, count+1)
		for i := 0; i <= count; i++ {
			loopRange[i] = i
		}
		rangeElems = loopRange
		logger.Debug().Str("count", fmt.Sprintf("%d", count)).Msg("count processed")
	}
	return rangeElems, nil
}

func paddedReplace(num, padding int) string {
	if padding > 0 {
		return fmt.Sprintf("%0*d", padding+1, num)
	}
	return strconv.Itoa(num)
}

func ProcessCommands(loopRange []int, command string, padding int) error {
	logger := utils.GetLogger("loopcmd")
	for _, num := range loopRange {
		cmdToRun := strings.ReplaceAll(command, "$i", paddedReplace(num, padding))
		logger.Debug().Str("command", cmdToRun).Msg("executing command")
		cmd := exec.Command("sh", "-c", cmdToRun)
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err := cmd.Run()
		if err != nil {
			logger.Error().Err(err).Str("stderr", stderrBuf.String()).Msg("failure")
			return fmt.Errorf("command execution failed: %w", err)
		}
		if stdoutBuf.Len() > 0 {
			fmt.Println(utils.OutDebug(stdoutBuf.String()))
		}
	}
	return nil
}
