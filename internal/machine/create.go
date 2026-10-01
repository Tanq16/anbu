package machine

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/scaffold"
	"github.com/tanq16/anbu/internal/vault"
)

type CreateConfig struct {
	Name         string
	InstanceType string
	Arch         string
	VCPU         int
	MemoryGB     int
	DiskGB       int
	Timezone     string
}

func (cfg CreateConfig) validate() error {
	switch {
	case cfg.Name == "":
		return errors.New("machine name is required")
	case cfg.InstanceType == "":
		return errors.New("instance type is required")
	case cfg.DiskGB <= 0:
		return fmt.Errorf("disk size must be positive, got %d", cfg.DiskGB)
	}
	_, err := amiParameter(cfg.Arch)
	return err
}

func Create(ctx context.Context, c *awsx.Clients, keys *vault.Store, cfg CreateConfig) (*Info, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	existing, err := c.FindInstance(ctx, cfg.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("a machine named %q already exists as %s (%s)", cfg.Name, existing.ID, existing.State)
	}

	found, err := findScaffolding(ctx, c, keys)
	if err != nil {
		return nil, err
	}

	ami, err := resolveAMI(ctx, c, cfg.Arch)
	if err != nil {
		return nil, err
	}
	timezone := cfg.Timezone
	if timezone == "" {
		timezone = LocalTimezone()
	}
	userData := base64.StdEncoding.EncodeToString([]byte(renderUserData(timezone)))

	out, err := c.EC2.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:          aws.String(ami),
		InstanceType:     ec2types.InstanceType(cfg.InstanceType),
		MinCount:         aws.Int32(1),
		MaxCount:         aws.Int32(1),
		KeyName:          aws.String(found.keyPairName),
		SubnetId:         aws.String(found.subnetID),
		SecurityGroupIds: []string{found.securityGroupID},
		BlockDeviceMappings: []ec2types.BlockDeviceMapping{{
			DeviceName: aws.String(rootDevice),
			Ebs: &ec2types.EbsBlockDevice{
				VolumeType:          ec2types.VolumeTypeGp3,
				VolumeSize:          aws.Int32(int32(cfg.DiskGB)),
				Encrypted:           aws.Bool(true),
				DeleteOnTermination: aws.Bool(true),
			},
		}},
		MetadataOptions: &ec2types.InstanceMetadataOptionsRequest{
			HttpTokens:              ec2types.HttpTokensStateRequired,
			HttpPutResponseHopLimit: aws.Int32(2),
			InstanceMetadataTags:    ec2types.InstanceMetadataTagsStateEnabled,
		},
		TagSpecifications: []ec2types.TagSpecification{
			awsx.TagSpecs(ec2types.ResourceTypeInstance, cfg.Name),
			awsx.TagSpecs(ec2types.ResourceTypeVolume, cfg.Name),
		},
		UserData: aws.String(userData),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Instances) == 0 {
		return nil, errors.New("ec2:RunInstances returned no instance")
	}
	id := aws.ToString(out.Instances[0].InstanceId)

	if err := waitRunning(ctx, c, id); err != nil {
		return nil, err
	}
	ip, err := publicIP(ctx, c, id)
	if err != nil {
		return nil, err
	}

	return &Info{
		Name:         cfg.Name,
		InstanceID:   id,
		InstanceType: cfg.InstanceType,
		Arch:         cfg.Arch,
		VCPU:         cfg.VCPU,
		MemoryGB:     cfg.MemoryGB,
		DiskGB:       cfg.DiskGB,
		PublicIP:     ip,
		State:        string(ec2types.InstanceStateNameRunning),
	}, nil
}

type scaffolding struct {
	subnetID        string
	securityGroupID string
	keyPairName     string
}

func findScaffolding(ctx context.Context, c *awsx.Clients, keys *vault.Store) (*scaffolding, error) {
	var found scaffolding
	lookups := []struct {
		kind  string
		field *string
		find  func(context.Context) (string, error)
	}{
		{"subnet", &found.subnetID, c.FindSubnet},
		{"security group", &found.securityGroupID, c.FindSecurityGroup},
	}
	for _, lookup := range lookups {
		id, err := lookup.find(ctx)
		if err != nil {
			return nil, err
		}
		if id == "" {
			return nil, fmt.Errorf("no managed %s in %s, run scaffold setup", lookup.kind, c.Region)
		}
		*lookup.field = id
	}

	key, err := scaffold.CheckKey(ctx, c, keys)
	if err != nil {
		return nil, err
	}
	if key.State == scaffold.KeyNone {
		return nil, fmt.Errorf("no managed key pair in %s, run scaffold setup", c.Region)
	}
	if err := key.Err(c.Region); err != nil {
		return nil, err
	}
	found.keyPairName = key.KeyPairName()
	return &found, nil
}
