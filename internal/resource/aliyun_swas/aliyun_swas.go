package aliyun_swas

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	swasopen20200601 "github.com/alibabacloud-go/swas-open-20200601/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"os"
)

type AliyunSwas struct {
	client *swasopen20200601.Client
}

func (a *AliyunSwas) getClient() *swasopen20200601.Client {
	if a.client == nil {
		config := &openapi.Config{
			AccessKeyId:     tea.String(os.Getenv("ALIYUN_ACCESS_KEY_ID")),
			AccessKeySecret: tea.String(os.Getenv("ALIYUN_ACCESS_KEY_SECRET")),
			RegionId:        tea.String(a.GetRegion()),
			Endpoint:        tea.String(fmt.Sprintf("swas.%s.aliyuncs.com", a.GetRegion())),
		}

		client, err := swasopen20200601.NewClient(config)

		if err != nil {
			panic("create aliyun swas client error:" + err.Error())
		}

		a.client = client
	}

	return a.client
}

func (a *AliyunSwas) GetInstances() ([]resource.InstanceInfo, error) {
	return a.getInstances()
}

func (a *AliyunSwas) getInstances() ([]resource.InstanceInfo, error) {
	list := make([]resource.InstanceInfo, 0)

	page := 1

	for {
		result, err := a.getClient().ListInstances(&swasopen20200601.ListInstancesRequest{
			RegionId:   tea.String(a.GetRegion()),
			Status:     tea.String("Running"),
			PageSize:   tea.Int32(100),
			PageNumber: tea.Int32(int32(page)),
		})

		if err != nil {
			return nil, err
		}

		for _, item := range result.Body.Instances {
			ipAddress := ""

			if a.GetIPType() == "public" {
				ipAddress = *item.PublicIpAddress
			} else {
				ipAddress = *item.InnerIpAddress
			}

			list = append(list, resource.InstanceInfo{
				Type:          resource.EcsInstanceType,
				ID:            *item.InstanceId,
				PublicAddress: ipAddress,
				PublicPort:    9100,
			})
		}

		if len(result.Body.Instances) == 100 {
			page++
			continue
		}

		break
	}

	return list, nil
}

func (a *AliyunSwas) GetIPType() string {
	if os.Getenv("ALIYUN_PUBLIC_IP") == "0" {
		return "private"
	}

	return "public"
}

func (a *AliyunSwas) GetRegion() string {
	return os.Getenv("ALIYUN_REGION")
}
