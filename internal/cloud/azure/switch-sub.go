package azure

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	u "github.com/tanq16/anbu/utils"
)

type Subscription struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func SwitchSubscription() error {
	cmd := exec.Command("az", "account", "list", "--query", "[].{name:name,id:id}", "-o", "json", "--all")
	var listStderr strings.Builder
	cmd.Stderr = &listStderr
	output, err := cmd.Output()
	if err != nil {
		detail := strings.TrimSpace(listStderr.String())
		if detail != "" {
			err = fmt.Errorf("%s: %w", detail, err)
		}
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}
	var subscriptions []Subscription
	if err := json.Unmarshal(output, &subscriptions); err != nil {
		return fmt.Errorf("failed to parse subscriptions: %w", err)
	}
	if len(subscriptions) == 0 {
		return fmt.Errorf("no subscriptions found")
	}
	options := make([]string, len(subscriptions))
	for i, sub := range subscriptions {
		options[i] = fmt.Sprintf("%s (%s)", sub.Name, sub.ID)
	}
	idx, err := u.PromptSelect("Select subscription to activate:", options)
	if err != nil {
		return err
	}
	if idx < 0 {
		return nil
	}
	selectedSub := subscriptions[idx]
	setCmd := exec.Command("az", "account", "set", "--subscription", selectedSub.ID)
	var setStderr strings.Builder
	setCmd.Stderr = &setStderr
	if err := setCmd.Run(); err != nil {
		detail := strings.TrimSpace(setStderr.String())
		if detail != "" {
			err = fmt.Errorf("%s: %w", detail, err)
		}
		return fmt.Errorf("failed to set subscription: %w", err)
	}
	u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(selectedSub.Name), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Subscription switched (%s)", selectedSub.ID))))
	return nil
}
