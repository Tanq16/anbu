package awsx

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const (
	TagKey   = "ManagedBy"
	TagValue = "sharingan"

	pricingRegion = "us-east-1"
)

type Clients struct {
	EC2     *ec2.Client
	SSM     *ssm.Client
	STS     *sts.Client
	Pricing *pricing.Client
	Region  string
	Profile string
	Account string
	ARN     string
}

func New(ctx context.Context, cfg aws.Config) (*Clients, error) {
	pricingCfg := cfg.Copy()
	pricingCfg.Region = pricingRegion

	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return &Clients{
		EC2:     ec2.NewFromConfig(cfg),
		SSM:     ssm.NewFromConfig(cfg),
		STS:     stsClient,
		Pricing: pricing.NewFromConfig(pricingCfg),
		Region:  cfg.Region,
		Account: aws.ToString(identity.Account),
		ARN:     aws.ToString(identity.Arn),
	}, nil
}

func ManagedFilter() []ec2types.Filter {
	return []ec2types.Filter{{
		Name:   aws.String("tag:" + TagKey),
		Values: []string{TagValue},
	}}
}

func NameFilter(name string) ec2types.Filter {
	return ec2types.Filter{
		Name:   aws.String("tag:Name"),
		Values: []string{name},
	}
}

func TagSpecs(rt ec2types.ResourceType, name string) ec2types.TagSpecification {
	tags := []ec2types.Tag{{Key: aws.String(TagKey), Value: aws.String(TagValue)}}
	if name != "" {
		tags = append(tags, ec2types.Tag{Key: aws.String("Name"), Value: aws.String(name)})
	}
	return ec2types.TagSpecification{ResourceType: rt, Tags: tags}
}

func (c *Clients) HostKeyAlias(name string) string {
	return "sharingan-" + c.Account + "-" + c.Region + "-" + name
}
