package azure

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

type Subscription struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func SwitchSubscription() error {
	cmd := exec.Command("az", "account", "list", "--query", "[].{name:name,id:id}", "-o", "json", "--all")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}
	var subscriptions []Subscription
	if err := json.Unmarshal(output, &subscriptions); err != nil {
		return fmt.Errorf("failed to parse subscriptions: %w", err)
	}
	if len(subscriptions) == 0 {
		return fmt.Errorf("no subscriptions found")
	}
	table := u.NewTable([]string{"#", "Name", "ID"})
	for i, sub := range subscriptions {
		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("%d", i+1),
			sub.Name,
			u.FDebug(sub.ID),
		})
	}
	table.PrintTable(false)
	fmt.Println()
	fmt.Print(u.FInfo("Select subscription number to activate: "))
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
	setCmd := exec.Command("az", "account", "set", "--subscription", selectedSub.ID)
	if err := setCmd.Run(); err != nil {
		return fmt.Errorf("failed to set subscription: %w", err)
	}
	log.Info().Str("subscription", selectedSub.Name).Str("id", selectedSub.ID).Msg("subscription switched successfully")
	return nil
}
