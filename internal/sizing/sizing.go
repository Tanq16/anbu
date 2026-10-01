package sizing

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/pricing"
)

type Class string

const (
	Burstable Class = "burstable"
	Dedicated Class = "dedicated"
)

const (
	ArchX86 = "x86_64"
	ArchARM = "arm64"
)

const mibPerGB = 1024

type allowKey struct {
	arch  string
	class Class
}

var allowlist = map[allowKey][]string{
	{ArchX86, Burstable}: {"t3a", "t3"},
	{ArchX86, Dedicated}: {"m7a", "m8a", "m7i", "c7a", "c8a", "c7i", "r7a", "r8a", "r7i"},
	{ArchARM, Burstable}: {"t4g"},
	{ArchARM, Dedicated}: {"m9g", "m8g", "m7g", "c9g", "c8g", "c7g", "r8g", "r7g"},
}

type Shape struct {
	VCPU     int
	MemoryGB int
}

func Shapes() []Shape {
	return []Shape{
		{1, 2}, {1, 4},
		{2, 4}, {2, 8}, {2, 16},
		{4, 8}, {4, 16}, {4, 32},
		{8, 16}, {8, 32}, {8, 64},
		{16, 32}, {16, 64}, {16, 128},
	}
}

func DiskSizes() []int {
	return []int{50, 100, 120, 150, 200, 300, 500}
}

type Resolved struct {
	InstanceType string
	VCPU         int
	MemoryGB     int
	Arch         string
	Class        Class
	HourlyUSD    float64
}

type priceTable interface {
	ComputeHourly(instanceType string) (float64, bool)
}

type typeInfo struct {
	name      string
	vcpu      int
	memoryMiB int64
	arches    []string
}

type Resolver struct {
	ec2    *ec2.Client
	prices priceTable
	types  []typeInfo
}

func NewResolver(ctx context.Context, c *awsx.Clients, prices *pricing.Table) (*Resolver, error) {
	offered, err := regionOfferings(ctx, c)
	if err != nil {
		return nil, err
	}
	described, err := describeFamilies(ctx, c.EC2, allowedFamilies())
	if err != nil {
		return nil, err
	}

	types := make([]typeInfo, 0, len(described))
	for _, info := range described {
		if _, ok := offered[info.name]; ok {
			types = append(types, info)
		}
	}
	slices.SortFunc(types, func(a, b typeInfo) int { return cmp.Compare(a.name, b.name) })
	return &Resolver{ec2: c.EC2, prices: prices, types: types}, nil
}

func (r *Resolver) Resolve(shape Shape, arch string, class Class) (*Resolved, error) {
	families, ok := allowlist[allowKey{arch, class}]
	if !ok {
		return nil, fmt.Errorf("no instance families allowed for arch %s class %s", arch, class)
	}
	var best *Resolved
	for _, info := range r.types {
		if !slices.Contains(families, familyOf(info.name)) {
			continue
		}
		if info.vcpu != shape.VCPU || info.memoryMiB != int64(shape.MemoryGB)*mibPerGB {
			continue
		}
		if !slices.Contains(info.arches, arch) {
			continue
		}
		hourly, priced := r.prices.ComputeHourly(info.name)
		if !priced {
			continue
		}
		if best != nil && hourly >= best.HourlyUSD {
			continue
		}
		best = &Resolved{
			InstanceType: info.name,
			VCPU:         info.vcpu,
			MemoryGB:     int(info.memoryMiB / mibPerGB),
			Arch:         arch,
			Class:        class,
			HourlyUSD:    hourly,
		}
	}
	return best, nil
}

func (r *Resolver) Details(ctx context.Context, instanceType string) (vcpu, memoryGB int, arches []string, err error) {
	if i := slices.IndexFunc(r.types, func(info typeInfo) bool { return info.name == instanceType }); i >= 0 {
		info := r.types[i]
		return info.vcpu, int(info.memoryMiB / mibPerGB), info.arches, nil
	}
	described, err := describeTypes(ctx, r.ec2, &ec2.DescribeInstanceTypesInput{
		InstanceTypes: []ec2types.InstanceType{ec2types.InstanceType(instanceType)},
	})
	if err != nil {
		return 0, 0, nil, err
	}
	if len(described) == 0 {
		return 0, 0, nil, fmt.Errorf("no instance type named %s", instanceType)
	}
	info := described[0]
	return info.vcpu, int(info.memoryMiB / mibPerGB), info.arches, nil
}

func familyOf(instanceType string) string {
	family, _, _ := strings.Cut(instanceType, ".")
	return family
}

func allowedFamilies() []string {
	var families []string
	for _, group := range allowlist {
		for _, family := range group {
			if !slices.Contains(families, family) {
				families = append(families, family)
			}
		}
	}
	slices.Sort(families)
	return families
}

func regionOfferings(ctx context.Context, c *awsx.Clients) (map[string]struct{}, error) {
	offered := make(map[string]struct{})
	pager := ec2.NewDescribeInstanceTypeOfferingsPaginator(c.EC2, &ec2.DescribeInstanceTypeOfferingsInput{
		LocationType: ec2types.LocationTypeRegion,
	})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, offering := range page.InstanceTypeOfferings {
			offered[string(offering.InstanceType)] = struct{}{}
		}
	}
	if len(offered) == 0 {
		return nil, fmt.Errorf("no instance types offered in %s", c.Region)
	}
	return offered, nil
}

func describeFamilies(ctx context.Context, client *ec2.Client, families []string) ([]typeInfo, error) {
	patterns := make([]string, len(families))
	for i, family := range families {
		patterns[i] = family + ".*"
	}
	return describeTypes(ctx, client, &ec2.DescribeInstanceTypesInput{
		Filters: []ec2types.Filter{{Name: aws.String("instance-type"), Values: patterns}},
	})
}

func describeTypes(ctx context.Context, client *ec2.Client, in *ec2.DescribeInstanceTypesInput) ([]typeInfo, error) {
	var infos []typeInfo
	pager := ec2.NewDescribeInstanceTypesPaginator(client, in)
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, it := range page.InstanceTypes {
			if it.VCpuInfo == nil || it.MemoryInfo == nil || it.ProcessorInfo == nil {
				continue
			}
			arches := make([]string, 0, len(it.ProcessorInfo.SupportedArchitectures))
			for _, arch := range it.ProcessorInfo.SupportedArchitectures {
				arches = append(arches, string(arch))
			}
			infos = append(infos, typeInfo{
				name:      string(it.InstanceType),
				vcpu:      int(aws.ToInt32(it.VCpuInfo.DefaultVCpus)),
				memoryMiB: aws.ToInt64(it.MemoryInfo.SizeInMiB),
				arches:    arches,
			})
		}
	}
	return infos, nil
}

func Classes() []Class {
	return []Class{Burstable, Dedicated}
}

func Arches() []string {
	return []string{ArchARM, ArchX86}
}

func (s Shape) ID() string {
	return fmt.Sprintf("%dvcpu-%dgb", s.VCPU, s.MemoryGB)
}

func ParseShape(id string) (Shape, bool) {
	i := slices.IndexFunc(Shapes(), func(s Shape) bool { return s.ID() == id })
	if i < 0 {
		return Shape{}, false
	}
	return Shapes()[i], true
}

func ClassOf(instanceType string) Class {
	if strings.HasPrefix(familyOf(instanceType), "t") {
		return Burstable
	}
	return Dedicated
}
