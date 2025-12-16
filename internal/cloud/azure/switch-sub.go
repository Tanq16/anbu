package azure

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func SwitchSubscription() error {
	cmd := exec.Command("az", "account", "list", "--query", "[].name", "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var subscriptions []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			subscriptions = append(subscriptions, trimmed)
		}
	}
	if len(subscriptions) == 0 {
		return fmt.Errorf("no subscriptions found")
	}
	fmt.Println("Azure Subscriptions:")
	for i, sub := range subscriptions {
		fmt.Printf("%d. %s\n", i+1, sub)
	}
	fmt.Println()
	fmt.Print("Select subscription number to activate: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	subNumber, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return fmt.Errorf("invalid subscription number: %w", err)
	}
	if subNumber < 1 || subNumber > len(subscriptions) {
		return fmt.Errorf("subscription number out of range")
	}
	selectedSub := subscriptions[subNumber-1]
	setCmd := exec.Command("az", "account", "set", "--subscription", selectedSub)
	if err := setCmd.Run(); err != nil {
		return fmt.Errorf("failed to set subscription: %w", err)
	}
	log.Info().Str("subscription", selectedSub).Msg("subscription switched successfully")
	return nil
}
