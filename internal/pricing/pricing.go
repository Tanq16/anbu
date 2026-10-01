package pricing

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awspricing "github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingtypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/tanq16/anbu/internal/awsx"
)

const hoursPerMonth = 730

type Table struct {
	compute    map[string]float64
	gp3GBMonth float64
	ipv4Hourly float64
}

func Fetch(ctx context.Context, c *awsx.Clients) (*Table, error) {
	compute, err := sweep(ctx, c)
	if err != nil {
		return nil, err
	}
	gp3, err := lookupRate(ctx, c, "AmazonEC2", "gp3 storage", []pricingtypes.Filter{
		termMatch("regionCode", c.Region),
		termMatch("volumeApiName", "gp3"),
		termMatch("productFamily", "Storage"),
	})
	if err != nil {
		return nil, err
	}
	ipv4, err := lookupRate(ctx, c, "AmazonVPC", "in-use public IPv4", []pricingtypes.Filter{
		termMatch("regionCode", c.Region),
		termMatch("group", "VPCPublicIPv4Address"),
		termMatch("groupDescription", "Hourly charge for In-use Public IPv4 Addresses"),
	})
	if err != nil {
		return nil, err
	}
	return &Table{compute: compute, gp3GBMonth: gp3, ipv4Hourly: ipv4}, nil
}

func (t *Table) ComputeHourly(instanceType string) (float64, bool) {
	rate, ok := t.compute[instanceType]
	return rate, ok
}

func (t *Table) GP3GBMonth() float64 { return t.gp3GBMonth }

func (t *Table) PublicIPv4Hourly() float64 { return t.ipv4Hourly }

func (t *Table) RunningHourly(computeHourly float64, diskGB int) float64 {
	return computeHourly + float64(diskGB)*(t.gp3GBMonth/hoursPerMonth) + t.ipv4Hourly
}

func (t *Table) RunningMonth(computeHourly float64, diskGB int) float64 {
	return t.RunningHourly(computeHourly, diskGB) * hoursPerMonth
}

func (t *Table) StoppedMonth(diskGB int) float64 {
	return float64(diskGB) * t.gp3GBMonth
}

func sweep(ctx context.Context, c *awsx.Clients) (map[string]float64, error) {
	in := &awspricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters: []pricingtypes.Filter{
			termMatch("regionCode", c.Region),
			termMatch("operatingSystem", "Linux"),
			termMatch("tenancy", "Shared"),
			termMatch("preInstalledSw", "NA"),
			termMatch("capacitystatus", "Used"),
			termMatch("marketoption", "OnDemand"),
		},
	}
	rates := make(map[string]float64)
	pager := awspricing.NewGetProductsPaginator(c.Pricing, in)
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, raw := range page.PriceList {
			e, err := parseEntry(raw)
			if err != nil {
				return nil, err
			}
			if e.instanceType == "" {
				continue
			}
			rates[e.instanceType] = e.usd
		}
	}
	if len(rates) == 0 {
		return nil, fmt.Errorf("no on-demand linux prices returned for %s", c.Region)
	}
	return rates, nil
}

func lookupRate(ctx context.Context, c *awsx.Clients, service, label string, filters []pricingtypes.Filter) (float64, error) {
	out, err := c.Pricing.GetProducts(ctx, &awspricing.GetProductsInput{
		ServiceCode: aws.String(service),
		Filters:     filters,
	})
	if err != nil {
		return 0, err
	}
	if len(out.PriceList) == 0 {
		return 0, fmt.Errorf("no price for %s in %s", label, c.Region)
	}
	e, err := parseEntry(out.PriceList[0])
	if err != nil {
		return 0, err
	}
	return e.usd, nil
}

func termMatch(field, value string) pricingtypes.Filter {
	return pricingtypes.Filter{
		Type:  pricingtypes.FilterTypeTermMatch,
		Field: aws.String(field),
		Value: aws.String(value),
	}
}

type entry struct {
	instanceType string
	usd          float64
}

type priceListEntry struct {
	Product struct {
		Attributes struct {
			InstanceType string `json:"instanceType"`
		} `json:"attributes"`
	} `json:"product"`
	Terms struct {
		OnDemand map[string]struct {
			PriceDimensions map[string]struct {
				PricePerUnit struct {
					USD string `json:"USD"`
				} `json:"pricePerUnit"`
			} `json:"priceDimensions"`
		} `json:"OnDemand"`
	} `json:"terms"`
}

func parseEntry(raw string) (entry, error) {
	var parsed priceListEntry
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return entry{}, err
	}
	for _, term := range parsed.Terms.OnDemand {
		for _, dim := range term.PriceDimensions {
			usd, err := strconv.ParseFloat(dim.PricePerUnit.USD, 64)
			if err != nil {
				return entry{}, err
			}
			return entry{instanceType: parsed.Product.Attributes.InstanceType, usd: usd}, nil
		}
	}
	return entry{}, fmt.Errorf("price list entry for %q carries no on-demand rate", parsed.Product.Attributes.InstanceType)
}
