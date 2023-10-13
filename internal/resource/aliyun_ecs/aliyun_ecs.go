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
		client, err := ecs20140526.NewClient(&openapi.Config{
			AccessKeyId:     tea.String(os.Getenv("ALIYUN_ACCESS_KEY_ID")),
			AccessKeySecret: tea.String(os.Getenv("ALIYUN_ACCESS_KEY_SECRET")),
			RegionId:        tea.String(a.GetRegion()),
		})

		if err != nil {
			panic("init client error:" + err.Error())
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
			if len(item.PublicIpAddress.IpAddress) == 0 {
				continue
			}

			list = append(list, resource.InstanceInfo{
				Type:          resource.EcsInstanceType,
				ID:            *item.InstanceId,
				PublicAddress: *item.PublicIpAddress.IpAddress[0],
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

func (a *AliyunEcs) GetUsername() string {
	return os.Getenv("MYSQL_USERNAME")
}

func (a *AliyunEcs) GetPassword() string {
	return os.Getenv("MYSQL_PASSWORD")
}

func (a *AliyunEcs) GetRegion() string {
	return os.Getenv("ALIYUN_REGION")
}
