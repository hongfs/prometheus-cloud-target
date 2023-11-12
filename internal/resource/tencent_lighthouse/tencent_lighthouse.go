package tencent_lighthouse

import (
	"github.com/alibabacloud-go/tea/tea"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	"os"
)

type TencentLighthouse struct {
	client *lighthouse.Client
}

func (t *TencentLighthouse) getClient() *lighthouse.Client {
	if t.client == nil {
		credential := common.NewCredential(
			os.Getenv("ALIYUN_ACCESS_KEY_ID"),
			os.Getenv("ALIYUN_ACCESS_KEY_SECRET"),
		)

		client, err := lighthouse.NewClient(credential, t.GetRegion(), nil)

		if err != nil {
			panic("init client error:" + err.Error())
		}

		t.client = client
	}

	return t.client
}

func (a *TencentLighthouse) GetInstances() ([]resource.InstanceInfo, error) {
	return a.getInstances()
}

func (a *TencentLighthouse) getInstances() ([]resource.InstanceInfo, error) {
	list := make([]resource.InstanceInfo, 0)

	page := 1

	for {
		result, err := a.getClient().DescribeInstances(&lighthouse.DescribeInstancesRequest{
			Limit:  tea.Int64(100),
			Offset: tea.Int64(int64(page * 100)),
		})

		if err != nil {
			return nil, err
		}

		for _, item := range result.Response.InstanceSet {
			item := item

			ipAddress := ""

			if a.GetIPType() == "public" {
				if len(item.PublicAddresses) == 0 {
					continue
				}

				ipAddress = *item.PublicAddresses[0]
			} else {
				if len(item.PrivateAddresses) == 0 {
					continue
				}

				ipAddress = *item.PrivateAddresses[0]
			}

			list = append(list, resource.InstanceInfo{
				Type:          resource.EcsInstanceType,
				ID:            *item.InstanceId,
				PublicAddress: ipAddress,
				PublicPort:    9100,
			})
		}

		if len(result.Response.InstanceSet) == 100 {
			page++
			continue
		}

		break
	}

	return list, nil
}

func (t *TencentLighthouse) GetIPType() string {
	if os.Getenv("TENCENT_PUBLIC_IP") == "0" {
		return "private"
	}

	return "public"
}

func (t *TencentLighthouse) GetRegion() string {
	return os.Getenv("TENCENT_REGION")
}
