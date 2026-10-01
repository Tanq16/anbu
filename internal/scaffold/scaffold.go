package scaffold

import (
	"strings"

	"github.com/tanq16/anbu/internal/vault"
)

type Action string

const (
	Created  Action = "created"
	Existing Action = "existing"
	Deleted  Action = "deleted"
)

type Event struct {
	Action   Action `json:"action"`
	Resource string `json:"resource"`
	ID       string `json:"id"`
}

type recorder struct {
	events []Event
}

func (r *recorder) notify(action Action, resource, id string) {
	r.events = append(r.events, Event{Action: action, Resource: resource, ID: id})
}

const (
	resVPC        = "vpc"
	resIGW        = "internet gateway"
	resSubnet     = "subnet"
	resRouteTable = "route table"
	resSecGroup   = "security group"
	resKeyPair    = "key pair"
	resKeySecret  = "key secret"
)

const (
	nameVPC        = "sharingan-vpc"
	nameIGW        = "sharingan-igw"
	nameSubnet     = "sharingan-subnet"
	nameRouteTable = "sharingan-rtb"
	nameSecGroup   = "sharingan-sg"
	nameKeyPair    = "sharingan-key"
)

func KeyName(account, region string) string {
	return vault.ScaffoldKeyPrefix + account + "-" + region
}

func ParseKeyName(name string) (account, region string, ok bool) {
	rest, ok := strings.CutPrefix(name, vault.ScaffoldKeyPrefix)
	if !ok {
		return "", "", false
	}
	account, region, ok = strings.Cut(rest, "-")
	if !ok || account == "" || region == "" {
		return "", "", false
	}
	return account, region, true
}
