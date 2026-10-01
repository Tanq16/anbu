package machine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tanq16/anbu/internal/awsx"
)

type ModifyConfig struct {
	Name         string
	InstanceType string
	VCPU         int
	MemoryGB     int
	Arches       []string
}

func Modify(ctx context.Context, c *awsx.Clients, cfg ModifyConfig) (*Info, error) {
	inst, err := requireInstance(ctx, c, cfg.Name)
	if err != nil {
		return nil, err
	}
	if err := validateModify(cfg, inst); err != nil {
		return nil, err
	}

	if inst.State != string(ec2types.InstanceStateNameStopped) {
		if _, err := c.EC2.StopInstances(ctx, &ec2.StopInstancesInput{InstanceIds: []string{inst.ID}}); err != nil {
			return nil, err
		}
		if err := waitStopped(ctx, c, inst.ID); err != nil {
			return nil, err
		}
	}

	_, err = c.EC2.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
		InstanceId:   aws.String(inst.ID),
		InstanceType: &ec2types.AttributeValue{Value: aws.String(cfg.InstanceType)},
	})
	if err != nil {
		if restartErr := startInstance(ctx, c, inst.ID); restartErr != nil {
			return nil, errors.Join(err, restartErr)
		}
		return nil, fmt.Errorf("%w (restarted on %s)", err, inst.Type)
	}

	if err := startInstance(ctx, c, inst.ID); err != nil {
		return nil, err
	}
	if err := waitRunning(ctx, c, inst.ID); err != nil {
		return nil, err
	}
	ip, err := publicIP(ctx, c, inst.ID)
	if err != nil {
		return nil, err
	}
	return &Info{
		Name:         cfg.Name,
		InstanceID:   inst.ID,
		InstanceType: cfg.InstanceType,
		Arch:         inst.Arch,
		VCPU:         cfg.VCPU,
		MemoryGB:     cfg.MemoryGB,
		DiskGB:       inst.DiskGB,
		PublicIP:     ip,
		State:        string(ec2types.InstanceStateNameRunning),
	}, nil
}

func validateModify(cfg ModifyConfig, inst *awsx.Instance) error {
	switch {
	case cfg.InstanceType == "":
		return errors.New("target instance type is required")
	case cfg.VCPU <= 0 || cfg.MemoryGB <= 0:
		return fmt.Errorf("target shape must be positive, got %d vCPU and %d GB", cfg.VCPU, cfg.MemoryGB)
	case cfg.InstanceType == inst.Type:
		return fmt.Errorf("%s already runs on %s", cfg.Name, inst.Type)
	case !slices.Contains(cfg.Arches, inst.Arch):
		return fmt.Errorf("%s supports %s, but %s runs on %s: an architecture change is refused",
			cfg.InstanceType, strings.Join(cfg.Arches, ", "), cfg.Name, inst.Arch)
	}
	return nil
}
