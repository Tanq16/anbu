package scaffold

import (
	"context"

	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/vault"
)

type Resource struct {
	Resource string `json:"resource"`
	ID       string `json:"id"`
}

type Status struct {
	Profile   string     `json:"profile"`
	Account   string     `json:"account"`
	Region    string     `json:"region"`
	Ready     bool       `json:"ready"`
	Resources []Resource `json:"resources"`
	Key       KeyStatus  `json:"key"`
}

func Inspect(ctx context.Context, c *awsx.Clients, keys *vault.Store) (Status, error) {
	key, err := CheckKey(ctx, c, keys)
	if err != nil {
		return Status{}, err
	}
	lookups := []struct {
		resource string
		find     func(context.Context) (string, error)
	}{
		{resVPC, c.FindVPC},
		{resIGW, c.FindIGW},
		{resSubnet, c.FindSubnet},
		{resRouteTable, c.FindRouteTable},
		{resSecGroup, c.FindSecurityGroup},
	}
	status := Status{Profile: c.Profile, Account: c.Account, Region: c.Region, Ready: true, Key: key}
	for _, lookup := range lookups {
		id, err := lookup.find(ctx)
		if err != nil {
			return Status{}, err
		}
		status.Resources = append(status.Resources, Resource{Resource: lookup.resource, ID: id})
		status.Ready = status.Ready && id != ""
	}
	var keyPairName string
	if key.keyPair != nil {
		keyPairName = key.keyPair.Name
	}
	status.Resources = append(status.Resources, Resource{Resource: resKeyPair, ID: keyPairName})
	status.Ready = status.Ready && keyPairName != ""
	return status, nil
}
