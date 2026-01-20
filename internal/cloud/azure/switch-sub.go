package azure

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

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
	u.LineBreak()
	input := u.InputWithClear("Select subscription number to activate: ")
	subNumber, err := strconv.Atoi(input)
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
	u.PrintSuccess(fmt.Sprintf("Subscription switched successfully: %s (%s)", u.FInfo(selectedSub.Name), u.FDebug(selectedSub.ID)))
	return nil
}
