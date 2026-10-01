package awsx

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Instance struct {
	ID       string
	Name     string
	Type     string
	Arch     string
	State    string
	PublicIP string
	DiskGB   int
	Launched time.Time
}

type KeyPair struct {
	Name      string
	PublicKey string
}

func (c *Clients) FindVPC(ctx context.Context) (string, error) {
	out, err := c.EC2.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{Filters: ManagedFilter()})
	if err != nil {
		return "", err
	}
	return single("vpc", collect(out.Vpcs, func(v ec2types.Vpc) string { return aws.ToString(v.VpcId) }))
}

func (c *Clients) FindSubnet(ctx context.Context) (string, error) {
	out, err := c.EC2.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{Filters: ManagedFilter()})
	if err != nil {
		return "", err
	}
	return single("subnet", collect(out.Subnets, func(s ec2types.Subnet) string { return aws.ToString(s.SubnetId) }))
}

func (c *Clients) FindIGW(ctx context.Context) (string, error) {
	out, err := c.EC2.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{Filters: ManagedFilter()})
	if err != nil {
		return "", err
	}
	return single("internet gateway", collect(out.InternetGateways, func(g ec2types.InternetGateway) string {
		return aws.ToString(g.InternetGatewayId)
	}))
}

func (c *Clients) FindRouteTable(ctx context.Context) (string, error) {
	out, err := c.EC2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{Filters: ManagedFilter()})
	if err != nil {
		return "", err
	}
	return single("route table", collect(out.RouteTables, func(r ec2types.RouteTable) string {
		return aws.ToString(r.RouteTableId)
	}))
}

func (c *Clients) FindSecurityGroup(ctx context.Context) (string, error) {
	out, err := c.EC2.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{Filters: ManagedFilter()})
	if err != nil {
		return "", err
	}
	return single("security group", collect(out.SecurityGroups, func(g ec2types.SecurityGroup) string {
		return aws.ToString(g.GroupId)
	}))
}

func (c *Clients) FindKeyPair(ctx context.Context) (*KeyPair, error) {
	out, err := c.EC2.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		Filters:          ManagedFilter(),
		IncludePublicKey: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	name, err := single("key pair", collect(out.KeyPairs, func(k ec2types.KeyPairInfo) string { return aws.ToString(k.KeyName) }))
	if err != nil || name == "" {
		return nil, err
	}
	for _, k := range out.KeyPairs {
		if aws.ToString(k.KeyName) == name {
			return &KeyPair{Name: name, PublicKey: aws.ToString(k.PublicKey)}, nil
		}
	}
	return nil, nil
}

func (c *Clients) ManagedInstances(ctx context.Context) ([]Instance, error) {
	return c.describeInstances(ctx, ManagedFilter())
}

func (c *Clients) FindInstance(ctx context.Context, name string) (*Instance, error) {
	found, err := c.describeInstances(ctx, append(ManagedFilter(), NameFilter(name)))
	if err != nil {
		return nil, err
	}
	if _, err := single(fmt.Sprintf("instance named %q", name), collect(found, func(i Instance) string { return i.ID })); err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	return &found[0], nil
}

func (c *Clients) describeInstances(ctx context.Context, filters []ec2types.Filter) ([]Instance, error) {
	pager := ec2.NewDescribeInstancesPaginator(c.EC2, &ec2.DescribeInstancesInput{Filters: filters})
	var instances []Instance
	roots := map[string]string{}
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, reservation := range page.Reservations {
			for _, inst := range reservation.Instances {
				var state string
				if inst.State != nil {
					state = string(inst.State.Name)
				}
				if state == string(ec2types.InstanceStateNameTerminated) {
					continue
				}
				id := aws.ToString(inst.InstanceId)
				instances = append(instances, Instance{
					ID:       id,
					Name:     tagValue(inst.Tags, "Name"),
					Type:     string(inst.InstanceType),
					Arch:     string(inst.Architecture),
					State:    state,
					PublicIP: aws.ToString(inst.PublicIpAddress),
					Launched: aws.ToTime(inst.LaunchTime),
				})
				if volume := rootVolumeID(inst); volume != "" {
					roots[id] = volume
				}
			}
		}
	}
	if len(roots) == 0 {
		return instances, nil
	}

	sizes, err := c.volumeSizes(ctx, slices.Collect(maps.Values(roots)))
	if err != nil {
		return nil, err
	}
	for i := range instances {
		instances[i].DiskGB = sizes[roots[instances[i].ID]]
	}
	return instances, nil
}

func (c *Clients) volumeSizes(ctx context.Context, volumeIDs []string) (map[string]int, error) {
	sizes := make(map[string]int, len(volumeIDs))
	pager := ec2.NewDescribeVolumesPaginator(c.EC2, &ec2.DescribeVolumesInput{VolumeIds: volumeIDs})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, volume := range page.Volumes {
			sizes[aws.ToString(volume.VolumeId)] = int(aws.ToInt32(volume.Size))
		}
	}
	return sizes, nil
}

func rootVolumeID(inst ec2types.Instance) string {
	root := aws.ToString(inst.RootDeviceName)
	var first string
	for _, mapping := range inst.BlockDeviceMappings {
		if mapping.Ebs == nil {
			continue
		}
		id := aws.ToString(mapping.Ebs.VolumeId)
		if aws.ToString(mapping.DeviceName) == root {
			return id
		}
		if first == "" {
			first = id
		}
	}
	return first
}

func tagValue(tags []ec2types.Tag, key string) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

func collect[T any](items []T, id func(T) string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if v := id(item); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func single(kind string, found []string) (string, error) {
	switch len(found) {
	case 0:
		return "", nil
	case 1:
		return found[0], nil
	default:
		return "", fmt.Errorf("expected one managed %s, found %d: %s", kind, len(found), strings.Join(found, ", "))
	}
}
