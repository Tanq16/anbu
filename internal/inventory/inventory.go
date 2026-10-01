package inventory

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tanq16/anbu/internal/awsx"
	"github.com/tanq16/anbu/internal/pricing"
	"github.com/tanq16/anbu/internal/sizing"
)

type Machine struct {
	Profile        string       `json:"profile"`
	Account        string       `json:"account"`
	Region         string       `json:"region"`
	Name           string       `json:"name"`
	InstanceID     string       `json:"instance_id"`
	InstanceType   string       `json:"type"`
	Class          sizing.Class `json:"class"`
	Arch           string       `json:"arch"`
	VCPU           int          `json:"vcpu"`
	MemoryGB       int          `json:"memory"`
	DiskGB         int          `json:"disk"`
	PublicIP       string       `json:"public_ip"`
	State          string       `json:"state"`
	Created        time.Time    `json:"created"`
	Hourly         float64      `json:"hourly"`
	MonthlyRunning float64      `json:"monthly_running"`
	MonthlyStopped float64      `json:"monthly_stopped"`
	Alias          string       `json:"alias"`
}

type shape struct {
	vcpu     int
	memoryGB int
}

func FromAPI(ctx context.Context, c *awsx.Clients, prices *pricing.Table) ([]Machine, error) {
	instances, err := c.ManagedInstances(ctx)
	if err != nil {
		return nil, err
	}
	shapes, err := instanceShapes(ctx, c, instances)
	if err != nil {
		return nil, err
	}
	machines := make([]Machine, 0, len(instances))
	for _, instance := range instances {
		s := shapes[instance.Type]
		m := Machine{
			Profile:      c.Profile,
			Account:      c.Account,
			Region:       c.Region,
			Name:         instance.Name,
			InstanceID:   instance.ID,
			InstanceType: instance.Type,
			Class:        sizing.ClassOf(instance.Type),
			Arch:         instance.Arch,
			VCPU:         s.vcpu,
			MemoryGB:     s.memoryGB,
			DiskGB:       instance.DiskGB,
			PublicIP:     instance.PublicIP,
			State:        instance.State,
			Created:      instance.Launched,
			Alias:        c.HostKeyAlias(instance.Name),
		}
		if prices != nil {
			m.Hourly, _ = prices.ComputeHourly(instance.Type)
			m.MonthlyRunning = prices.RunningMonth(m.Hourly, m.DiskGB)
			m.MonthlyStopped = prices.StoppedMonth(m.DiskGB)
		}
		machines = append(machines, m)
	}
	slices.SortFunc(machines, func(a, b Machine) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.InstanceID, b.InstanceID))
	})
	return machines, nil
}

func instanceShapes(ctx context.Context, c *awsx.Clients, instances []awsx.Instance) (map[string]shape, error) {
	var types []ec2types.InstanceType
	for _, instance := range instances {
		t := ec2types.InstanceType(instance.Type)
		if instance.Type != "" && !slices.Contains(types, t) {
			types = append(types, t)
		}
	}
	shapes := make(map[string]shape, len(types))
	if len(types) == 0 {
		return shapes, nil
	}

	pager := ec2.NewDescribeInstanceTypesPaginator(c.EC2, &ec2.DescribeInstanceTypesInput{InstanceTypes: types})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, info := range page.InstanceTypes {
			var s shape
			if info.VCpuInfo != nil {
				s.vcpu = int(aws.ToInt32(info.VCpuInfo.DefaultVCpus))
			}
			if info.MemoryInfo != nil {
				s.memoryGB = int(aws.ToInt64(info.MemoryInfo.SizeInMiB) / 1024)
			}
			shapes[string(info.InstanceType)] = s
		}
	}
	return shapes, nil
}
