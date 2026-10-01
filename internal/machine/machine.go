package machine

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/tanq16/anbu/internal/awsx"
)

//go:embed userdata.sh
var bootstrapScript string

const (
	archX86 = "x86_64"
	archARM = "arm64"

	SSHUser       = "ubuntu"
	rootDevice    = "/dev/sda1"
	timezoneToken = "__WORKSTATION_TZ__"

	waitTimeout = 10 * time.Minute
)

var ErrNotFound = errors.New("machine not found")

type NotRunningError struct {
	Name  string
	State string
}

func (e *NotRunningError) Error() string {
	return fmt.Sprintf("%s is %s and has no public address", e.Name, e.State)
}

type Info struct {
	Name         string `json:"name"`
	InstanceID   string `json:"instance_id,omitempty"`
	InstanceType string `json:"type,omitempty"`
	Arch         string `json:"arch,omitempty"`
	VCPU         int    `json:"vcpu,omitzero"`
	MemoryGB     int    `json:"memory,omitzero"`
	DiskGB       int    `json:"disk,omitzero"`
	PublicIP     string `json:"public_ip"`
	State        string `json:"state"`
}

func amiParameter(arch string) (string, error) {
	var slug string
	switch arch {
	case archX86:
		slug = "amd64"
	case archARM:
		slug = archARM
	default:
		return "", fmt.Errorf("unsupported architecture %q, want %s or %s", arch, archX86, archARM)
	}
	return "/aws/service/canonical/ubuntu/server/26.04/stable/current/" + slug + "/hvm/ebs-gp3/ami-id", nil
}

func resolveAMI(ctx context.Context, c *awsx.Clients, arch string) (string, error) {
	name, err := amiParameter(arch)
	if err != nil {
		return "", err
	}
	out, err := c.SSM.GetParameter(ctx, &ssm.GetParameterInput{Name: aws.String(name)})
	if err != nil {
		return "", err
	}
	if out.Parameter == nil || aws.ToString(out.Parameter.Value) == "" {
		return "", fmt.Errorf("ssm:GetParameter %s returned no value", name)
	}
	return aws.ToString(out.Parameter.Value), nil
}

func renderUserData(timezone string) string {
	return strings.ReplaceAll(bootstrapScript, timezoneToken, timezone)
}

func requireInstance(ctx context.Context, c *awsx.Clients, name string) (*awsx.Instance, error) {
	inst, err := c.FindInstance(ctx, name)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, fmt.Errorf("%w: no machine named %q in %s", ErrNotFound, name, c.Region)
	}
	return inst, nil
}

func publicIP(ctx context.Context, c *awsx.Clients, instanceID string) (string, error) {
	out, err := c.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}})
	if err != nil {
		return "", err
	}
	for _, reservation := range out.Reservations {
		for _, inst := range reservation.Instances {
			return aws.ToString(inst.PublicIpAddress), nil
		}
	}
	return "", fmt.Errorf("ec2:DescribeInstances found no instance %s", instanceID)
}

func waitRunning(ctx context.Context, c *awsx.Clients, instanceID string) error {
	input := &ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}}
	return ec2.NewInstanceRunningWaiter(c.EC2).Wait(ctx, input, waitTimeout)
}

func waitStopped(ctx context.Context, c *awsx.Clients, instanceID string) error {
	input := &ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}}
	return ec2.NewInstanceStoppedWaiter(c.EC2).Wait(ctx, input, waitTimeout)
}

func waitTerminated(ctx context.Context, c *awsx.Clients, instanceID string) error {
	input := &ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}}
	return ec2.NewInstanceTerminatedWaiter(c.EC2).Wait(ctx, input, waitTimeout)
}
