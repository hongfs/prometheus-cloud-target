package aliyun_ecs

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"os"
)

type AliyunEcs struct {
	client *ecs20140526.Client
}

func (a *AliyunEcs) getClient() *ecs20140526.Client {
	if a.client == nil {
		config := &openapi.Config{
			AccessKeyId:     tea.String(os.Getenv("ALIYUN_ACCESS_KEY_ID")),
			AccessKeySecret: tea.String(os.Getenv("ALIYUN_ACCESS_KEY_SECRET")),
			RegionId:        tea.String(a.GetRegion()),
		}

		if a.GetIPType() == "private" {
			config.Network = tea.String("vpc")
		}

		client, err := ecs20140526.NewClient(config)

		if err != nil {
			panic("create aliyun ecs client error:" + err.Error())
		}

		a.client = client
	}

	return a.client
}

func (a *AliyunEcs) GetInstances() ([]resource.InstanceInfo, error) {
	return a.getInstances()
}

func (a *AliyunEcs) getInstances() ([]resource.InstanceInfo, error) {
	list := make([]resource.InstanceInfo, 0)

	var next *string

	for {
		result, err := a.getClient().DescribeInstances(&ecs20140526.DescribeInstancesRequest{
			RegionId:            tea.String(a.GetRegion()),
			InstanceNetworkType: tea.String("vpc"),
			Status:              tea.String("Running"),
			PageSize:            tea.Int32(100),
			Tag: []*ecs20140526.DescribeInstancesRequestTag{
				{
					Key:   tea.String("has_node_exporter"),
					Value: tea.String("1"),
				},
			},
			NextToken: next,
		})

		if err != nil {
			return nil, err
		}

		for _, item := range result.Body.Instances.Instance {
			ipAddress := ""

			if a.GetIPType() == "public" {
				if len(item.PublicIpAddress.IpAddress) == 0 {
					continue
				}

				ipAddress = *item.PublicIpAddress.IpAddress[0]
			} else {
				if len(item.VpcAttributes.PrivateIpAddress.IpAddress) == 0 {
					continue
				}

				ipAddress = *item.VpcAttributes.PrivateIpAddress.IpAddress[0]
			}

			list = append(list, resource.InstanceInfo{
				Type:          resource.EcsInstanceType,
				ID:            *item.InstanceId,
				PublicAddress: ipAddress,
				PublicPort:    9100,
			})
		}

		if len(result.Body.Instances.Instance) == 100 {
			next = result.Body.NextToken
			continue
		}

		break
	}

	return list, nil
}

func (a *AliyunEcs) GetIPType() string {
	if os.Getenv("ALIYUN_PUBLIC_IP") == "0" {
		return "private"
	}

	return "public"
}

func (a *AliyunEcs) GetRegion() string {
	return os.Getenv("ALIYUN_REGION")
}
