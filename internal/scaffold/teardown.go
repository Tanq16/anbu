package scaffold

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/vault"
)

type InstancesExistError struct {
	Instances []string
}

func (e *InstancesExistError) Error() string {
	return "managed instances still exist: " + strings.Join(e.Instances, ", ")
}

func Teardown(ctx context.Context, c *awsx.Clients, keys *vault.Store) ([]Event, error) {
	instances, err := c.ManagedInstances(ctx)
	if err != nil {
		return nil, err
	}
	if blocking := blockingInstances(instances); len(blocking) > 0 {
		return nil, &InstancesExistError{Instances: blocking}
	}

	cfg := &recorder{}
	steps := []func(context.Context, *recorder, *awsx.Clients) error{
		deleteSecurityGroup,
		deleteSubnet,
		deleteRouteTable,
		deleteIGW,
		deleteVPC,
		deleteKeyPair,
	}
	for _, step := range steps {
		if err := step(ctx, cfg, c); err != nil {
			return cfg.events, err
		}
	}
	fingerprint := ""
	if k, ok := keys.MachineKey(c.Account, c.Region); ok {
		fingerprint = k.Fingerprint()
	}
	deleted, err := keys.DeleteMachineKey(c.Account, c.Region)
	if err != nil {
		return cfg.events, err
	}
	if deleted {
		cfg.notify(Deleted, resScaffoldKey, fingerprint)
	}
	return cfg.events, nil
}

func deleteSecurityGroup(ctx context.Context, cfg *recorder, c *awsx.Clients) error {
	id, err := c.FindSecurityGroup(ctx)
	if err != nil || id == "" {
		return err
	}
	if _, err := c.EC2.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{GroupId: aws.String(id)}); err != nil {
		return err
	}
	cfg.notify(Deleted, resSecGroup, id)
	return nil
}

func deleteSubnet(ctx context.Context, cfg *recorder, c *awsx.Clients) error {
	id, err := c.FindSubnet(ctx)
	if err != nil || id == "" {
		return err
	}
	if _, err := c.EC2.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{SubnetId: aws.String(id)}); err != nil {
		return err
	}
	cfg.notify(Deleted, resSubnet, id)
	return nil
}

func deleteRouteTable(ctx context.Context, cfg *recorder, c *awsx.Clients) error {
	id, err := c.FindRouteTable(ctx)
	if err != nil || id == "" {
		return err
	}
	out, err := c.EC2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{RouteTableIds: []string{id}})
	if err != nil {
		return err
	}
	for _, associationID := range subnetAssociations(out.RouteTables) {
		if _, err := c.EC2.DisassociateRouteTable(ctx, &ec2.DisassociateRouteTableInput{
			AssociationId: aws.String(associationID),
		}); err != nil {
			return err
		}
	}
	if _, err := c.EC2.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{RouteTableId: aws.String(id)}); err != nil {
		return err
	}
	cfg.notify(Deleted, resRouteTable, id)
	return nil
}

func deleteIGW(ctx context.Context, cfg *recorder, c *awsx.Clients) error {
	id, err := c.FindIGW(ctx)
	if err != nil || id == "" {
		return err
	}
	out, err := c.EC2.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{InternetGatewayIds: []string{id}})
	if err != nil {
		return err
	}
	for _, vpcID := range attachedVPCs(out.InternetGateways) {
		if _, err := c.EC2.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(id),
			VpcId:             aws.String(vpcID),
		}); err != nil {
			return err
		}
	}
	if _, err := c.EC2.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(id),
	}); err != nil {
		return err
	}
	cfg.notify(Deleted, resIGW, id)
	return nil
}

func deleteVPC(ctx context.Context, cfg *recorder, c *awsx.Clients) error {
	id, err := c.FindVPC(ctx)
	if err != nil || id == "" {
		return err
	}
	if _, err := c.EC2.DeleteVpc(ctx, &ec2.DeleteVpcInput{VpcId: aws.String(id)}); err != nil {
		return err
	}
	cfg.notify(Deleted, resVPC, id)
	return nil
}

func deleteKeyPair(ctx context.Context, cfg *recorder, c *awsx.Clients) error {
	kp, err := c.FindKeyPair(ctx)
	if err != nil || kp == nil {
		return err
	}
	if _, err := c.EC2.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{KeyName: aws.String(kp.Name)}); err != nil {
		return err
	}
	cfg.notify(Deleted, resKeyPair, kp.Name)
	return nil
}

func blockingInstances(instances []awsx.Instance) []string {
	var blocking []string
	for _, instance := range instances {
		if instance.State == string(ec2types.InstanceStateNameTerminated) {
			continue
		}
		if instance.Name == "" {
			blocking = append(blocking, fmt.Sprintf("%s (%s)", instance.ID, instance.State))
			continue
		}
		blocking = append(blocking, fmt.Sprintf("%s (%s, %s)", instance.Name, instance.ID, instance.State))
	}
	return blocking
}

func subnetAssociations(tables []ec2types.RouteTable) []string {
	var ids []string
	for _, table := range tables {
		for _, association := range table.Associations {
			if aws.ToBool(association.Main) {
				continue
			}
			if id := aws.ToString(association.RouteTableAssociationId); id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func attachedVPCs(gateways []ec2types.InternetGateway) []string {
	var ids []string
	for _, gateway := range gateways {
		for _, attachment := range gateway.Attachments {
			if id := aws.ToString(attachment.VpcId); id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}
