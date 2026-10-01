package machine

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tanq16/anbu/internal/awsx"
)

func Start(ctx context.Context, c *awsx.Clients, name string) (*Info, error) {
	inst, err := requireInstance(ctx, c, name)
	if err != nil {
		return nil, err
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
	return &Info{Name: name, InstanceID: inst.ID, InstanceType: inst.Type, Arch: inst.Arch, DiskGB: inst.DiskGB,
		PublicIP: ip, State: string(ec2types.InstanceStateNameRunning)}, nil
}

func Stop(ctx context.Context, c *awsx.Clients, name string) (*Info, error) {
	inst, err := requireInstance(ctx, c, name)
	if err != nil {
		return nil, err
	}
	if _, err := c.EC2.StopInstances(ctx, &ec2.StopInstancesInput{InstanceIds: []string{inst.ID}}); err != nil {
		return nil, err
	}
	if err := waitStopped(ctx, c, inst.ID); err != nil {
		return nil, err
	}
	return &Info{Name: name, InstanceID: inst.ID, InstanceType: inst.Type, Arch: inst.Arch, DiskGB: inst.DiskGB,
		State: string(ec2types.InstanceStateNameStopped)}, nil
}

func Remove(ctx context.Context, c *awsx.Clients, name string) (*Info, error) {
	inst, err := requireInstance(ctx, c, name)
	if err != nil {
		return nil, err
	}
	if _, err := c.EC2.TerminateInstances(ctx, &ec2.TerminateInstancesInput{InstanceIds: []string{inst.ID}}); err != nil {
		return nil, err
	}
	if err := waitTerminated(ctx, c, inst.ID); err != nil {
		return nil, err
	}
	return &Info{Name: name, InstanceID: inst.ID, State: string(ec2types.InstanceStateNameTerminated)}, nil
}

func startInstance(ctx context.Context, c *awsx.Clients, instanceID string) error {
	_, err := c.EC2.StartInstances(ctx, &ec2.StartInstancesInput{InstanceIds: []string{instanceID}})
	return err
}
