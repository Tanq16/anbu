package anbuGenerics

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	u "github.com/tanq16/anbu/utils"
)

func loopProcessRange(input string) ([]int, error) {
	var rangeElems []int
	if strings.Contains(input, "-") {
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
	} else {
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
	}
	return rangeElems, nil
}

func loopPaddedReplace(num, padding int) string {
	if padding > 0 {
		return fmt.Sprintf("%0*d", padding+1, num)
	}
	return strconv.Itoa(num)
}

func LoopProcessCommands(loopRangeStr string, command string, padding int) {
	loopRange, err := loopProcessRange(loopRangeStr)
	if err != nil {
		u.PrintError(fmt.Sprintf("Invalid range: %s", err))
		return
	}
	for _, num := range loopRange {
		cmdToRun := strings.ReplaceAll(command, "$i", loopPaddedReplace(num, padding))
		cmd := exec.Command("sh", "-c", cmdToRun)
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err := cmd.Run()
		if err != nil {
			u.PrintError(fmt.Sprintf("command execution failed: %s", err))
		}
		if stdoutBuf.Len() > 0 {
			u.PrintDebug(stdoutBuf.String())
		}
	}
}
