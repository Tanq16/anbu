package scaffold

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/vault"
)

const (
	vpcCIDR       = "10.20.0.0/16"
	subnetCIDR    = "10.20.1.0/24"
	anywhere      = "0.0.0.0/0"
	protoTCP      = "tcp"
	sgDescription = "sharingan workstation access"
)

var ingressPorts = []int32{22, 443}

func Setup(ctx context.Context, c *awsx.Clients, keys *vault.Store) ([]Event, error) {
	cfg := &recorder{}
	key, err := CheckKey(ctx, c, keys)
	if err != nil {
		return nil, err
	}
	if err := key.Err(c.Region); err != nil {
		return nil, err
	}
	vpcID, err := ensureVPC(ctx, cfg, c)
	if err != nil {
		return cfg.events, err
	}
	igwID, err := ensureIGW(ctx, cfg, c, vpcID)
	if err != nil {
		return cfg.events, err
	}
	subnetID, err := ensureSubnet(ctx, cfg, c, vpcID)
	if err != nil {
		return cfg.events, err
	}
	if _, err := ensureRouteTable(ctx, cfg, c, vpcID, igwID, subnetID); err != nil {
		return cfg.events, err
	}
	if _, err := ensureSecurityGroup(ctx, cfg, c, vpcID); err != nil {
		return cfg.events, err
	}
	err = ensureKeyPair(ctx, cfg, c, keys, key)
	return cfg.events, err
}

func ensureVPC(ctx context.Context, cfg *recorder, c *awsx.Clients) (string, error) {
	id, err := c.FindVPC(ctx)
	if err != nil {
		return "", err
	}
	action := Existing
	if id == "" {
		out, err := c.EC2.CreateVpc(ctx, &ec2.CreateVpcInput{
			CidrBlock:         aws.String(vpcCIDR),
			TagSpecifications: []ec2types.TagSpecification{awsx.TagSpecs(ec2types.ResourceTypeVpc, nameVPC)},
		})
		if err != nil {
			return "", err
		}
		if out.Vpc == nil {
			return "", fmt.Errorf("ec2:CreateVpc returned no vpc")
		}
		id = aws.ToString(out.Vpc.VpcId)
		action = Created
	}

	enabled := &ec2types.AttributeBooleanValue{Value: aws.Bool(true)}
	if _, err := c.EC2.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
		VpcId:            aws.String(id),
		EnableDnsSupport: enabled,
	}); err != nil {
		return "", err
	}
	if _, err := c.EC2.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
		VpcId:              aws.String(id),
		EnableDnsHostnames: enabled,
	}); err != nil {
		return "", err
	}
	cfg.notify(action, resVPC, id)
	return id, nil
}

func ensureIGW(ctx context.Context, cfg *recorder, c *awsx.Clients, vpcID string) (string, error) {
	id, err := c.FindIGW(ctx)
	if err != nil {
		return "", err
	}
	if id == "" {
		out, err := c.EC2.CreateInternetGateway(ctx, &ec2.CreateInternetGatewayInput{
			TagSpecifications: []ec2types.TagSpecification{awsx.TagSpecs(ec2types.ResourceTypeInternetGateway, nameIGW)},
		})
		if err != nil {
			return "", err
		}
		if out.InternetGateway == nil {
			return "", fmt.Errorf("ec2:CreateInternetGateway returned no internet gateway")
		}
		id = aws.ToString(out.InternetGateway.InternetGatewayId)
		if err := attachIGW(ctx, c, id, vpcID); err != nil {
			return "", err
		}
		cfg.notify(Created, resIGW, id)
		return id, nil
	}

	out, err := c.EC2.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{InternetGatewayIds: []string{id}})
	if err != nil {
		return "", err
	}
	if !attachedTo(out.InternetGateways, vpcID) {
		if err := attachIGW(ctx, c, id, vpcID); err != nil {
			return "", err
		}
	}
	cfg.notify(Existing, resIGW, id)
	return id, nil
}

func attachIGW(ctx context.Context, c *awsx.Clients, igwID, vpcID string) error {
	if _, err := c.EC2.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(igwID),
		VpcId:             aws.String(vpcID),
	}); err != nil {
		return err
	}
	return nil
}

func ensureSubnet(ctx context.Context, cfg *recorder, c *awsx.Clients, vpcID string) (string, error) {
	id, err := c.FindSubnet(ctx)
	if err != nil {
		return "", err
	}
	action := Existing
	if id == "" {
		zone, err := firstAvailableZone(ctx, c)
		if err != nil {
			return "", err
		}
		out, err := c.EC2.CreateSubnet(ctx, &ec2.CreateSubnetInput{
			VpcId:             aws.String(vpcID),
			CidrBlock:         aws.String(subnetCIDR),
			AvailabilityZone:  aws.String(zone),
			TagSpecifications: []ec2types.TagSpecification{awsx.TagSpecs(ec2types.ResourceTypeSubnet, nameSubnet)},
		})
		if err != nil {
			return "", err
		}
		if out.Subnet == nil {
			return "", fmt.Errorf("ec2:CreateSubnet returned no subnet")
		}
		id = aws.ToString(out.Subnet.SubnetId)
		action = Created
	}

	if _, err := c.EC2.ModifySubnetAttribute(ctx, &ec2.ModifySubnetAttributeInput{
		SubnetId:            aws.String(id),
		MapPublicIpOnLaunch: &ec2types.AttributeBooleanValue{Value: aws.Bool(true)},
	}); err != nil {
		return "", err
	}
	cfg.notify(action, resSubnet, id)
	return id, nil
}

func firstAvailableZone(ctx context.Context, c *awsx.Clients) (string, error) {
	out, err := c.EC2.DescribeAvailabilityZones(ctx, &ec2.DescribeAvailabilityZonesInput{
		Filters: []ec2types.Filter{{
			Name:   aws.String("state"),
			Values: []string{string(ec2types.AvailabilityZoneStateAvailable)},
		}},
	})
	if err != nil {
		return "", err
	}
	for _, zone := range out.AvailabilityZones {
		if name := aws.ToString(zone.ZoneName); name != "" {
			return name, nil
		}
	}
	return "", fmt.Errorf("no available availability zone in %s", c.Region)
}

func ensureRouteTable(ctx context.Context, cfg *recorder, c *awsx.Clients, vpcID, igwID, subnetID string) (string, error) {
	id, err := c.FindRouteTable(ctx)
	if err != nil {
		return "", err
	}
	if id == "" {
		out, err := c.EC2.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
			VpcId:             aws.String(vpcID),
			TagSpecifications: []ec2types.TagSpecification{awsx.TagSpecs(ec2types.ResourceTypeRouteTable, nameRouteTable)},
		})
		if err != nil {
			return "", err
		}
		if out.RouteTable == nil {
			return "", fmt.Errorf("ec2:CreateRouteTable returned no route table")
		}
		id = aws.ToString(out.RouteTable.RouteTableId)
		if err := createDefaultRoute(ctx, c, id, igwID); err != nil {
			return "", err
		}
		if err := associateSubnet(ctx, c, id, subnetID); err != nil {
			return "", err
		}
		cfg.notify(Created, resRouteTable, id)
		return id, nil
	}

	out, err := c.EC2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{RouteTableIds: []string{id}})
	if err != nil {
		return "", err
	}
	if !hasDefaultRoute(out.RouteTables) {
		if err := createDefaultRoute(ctx, c, id, igwID); err != nil {
			return "", err
		}
	}
	if !associatedWith(out.RouteTables, subnetID) {
		if err := associateSubnet(ctx, c, id, subnetID); err != nil {
			return "", err
		}
	}
	cfg.notify(Existing, resRouteTable, id)
	return id, nil
}

func createDefaultRoute(ctx context.Context, c *awsx.Clients, routeTableID, igwID string) error {
	if _, err := c.EC2.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:         aws.String(routeTableID),
		DestinationCidrBlock: aws.String(anywhere),
		GatewayId:            aws.String(igwID),
	}); err != nil {
		return err
	}
	return nil
}

func associateSubnet(ctx context.Context, c *awsx.Clients, routeTableID, subnetID string) error {
	if _, err := c.EC2.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(routeTableID),
		SubnetId:     aws.String(subnetID),
	}); err != nil {
		return err
	}
	return nil
}

func ensureSecurityGroup(ctx context.Context, cfg *recorder, c *awsx.Clients, vpcID string) (string, error) {
	id, err := c.FindSecurityGroup(ctx)
	if err != nil {
		return "", err
	}
	if id == "" {
		out, err := c.EC2.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
			GroupName:         aws.String(nameSecGroup),
			Description:       aws.String(sgDescription),
			VpcId:             aws.String(vpcID),
			TagSpecifications: []ec2types.TagSpecification{awsx.TagSpecs(ec2types.ResourceTypeSecurityGroup, nameSecGroup)},
		})
		if err != nil {
			return "", err
		}
		id = aws.ToString(out.GroupId)
		if err := authorizeIngress(ctx, c, id, ingressPorts); err != nil {
			return "", err
		}
		cfg.notify(Created, resSecGroup, id)
		return id, nil
	}

	out, err := c.EC2.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{GroupIds: []string{id}})
	if err != nil {
		return "", err
	}
	if missing := missingIngress(out.SecurityGroups, ingressPorts); len(missing) > 0 {
		if err := authorizeIngress(ctx, c, id, missing); err != nil {
			return "", err
		}
	}
	cfg.notify(Existing, resSecGroup, id)
	return id, nil
}

func authorizeIngress(ctx context.Context, c *awsx.Clients, groupID string, ports []int32) error {
	permissions := make([]ec2types.IpPermission, 0, len(ports))
	for _, port := range ports {
		permissions = append(permissions, ec2types.IpPermission{
			IpProtocol: aws.String(protoTCP),
			FromPort:   aws.Int32(port),
			ToPort:     aws.Int32(port),
			IpRanges:   []ec2types.IpRange{{CidrIp: aws.String(anywhere)}},
		})
	}
	if _, err := c.EC2.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       aws.String(groupID),
		IpPermissions: permissions,
	}); err != nil {
		return err
	}
	return nil
}

func ensureKeyPair(ctx context.Context, cfg *recorder, c *awsx.Clients, keys *vault.Store, key KeyStatus) error {
	stored := key.stored
	if stored == nil {
		sec, err := keys.GenerateSSHKey(key.Secret)
		if err != nil {
			return err
		}
		stored = &sec
		cfg.notify(Created, resKeySecret, sec.Name)
	} else {
		cfg.notify(Existing, resKeySecret, stored.Name)
	}
	if key.keyPair != nil {
		cfg.notify(Existing, resKeyPair, key.keyPair.Name)
		return nil
	}

	out, err := c.EC2.ImportKeyPair(ctx, &ec2.ImportKeyPairInput{
		KeyName:           aws.String(nameKeyPair),
		PublicKeyMaterial: []byte(stored.Fields["public_key"]),
		TagSpecifications: []ec2types.TagSpecification{awsx.TagSpecs(ec2types.ResourceTypeKeyPair, nameKeyPair)},
	})
	if err != nil {
		return err
	}
	cfg.notify(Created, resKeyPair, aws.ToString(out.KeyName))
	return nil
}

func attachedTo(gateways []ec2types.InternetGateway, vpcID string) bool {
	for _, gateway := range gateways {
		for _, attachment := range gateway.Attachments {
			if aws.ToString(attachment.VpcId) == vpcID {
				return true
			}
		}
	}
	return false
}

func hasDefaultRoute(tables []ec2types.RouteTable) bool {
	for _, table := range tables {
		for _, route := range table.Routes {
			if aws.ToString(route.DestinationCidrBlock) == anywhere {
				return true
			}
		}
	}
	return false
}

func associatedWith(tables []ec2types.RouteTable, subnetID string) bool {
	for _, table := range tables {
		for _, association := range table.Associations {
			if aws.ToString(association.SubnetId) == subnetID {
				return true
			}
		}
	}
	return false
}

func missingIngress(groups []ec2types.SecurityGroup, ports []int32) []int32 {
	var missing []int32
	for _, port := range ports {
		if !hasIngress(groups, port) {
			missing = append(missing, port)
		}
	}
	return missing
}

func hasIngress(groups []ec2types.SecurityGroup, port int32) bool {
	for _, group := range groups {
		for _, permission := range group.IpPermissions {
			if aws.ToString(permission.IpProtocol) != protoTCP {
				continue
			}
			if aws.ToInt32(permission.FromPort) != port || aws.ToInt32(permission.ToPort) != port {
				continue
			}
			for _, r := range permission.IpRanges {
				if aws.ToString(r.CidrIp) == anywhere {
					return true
				}
			}
		}
	}
	return false
}
