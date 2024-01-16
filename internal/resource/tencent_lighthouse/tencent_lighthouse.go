package tencent_lighthouse

import (
	"github.com/alibabacloud-go/tea/tea"
	"github.com/hongfs/prometheus-cloud-target/internal/common"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	auth "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

type TencentLighthouse struct {
	client *lighthouse.Client
}

func (t *TencentLighthouse) getClient() *lighthouse.Client {
	if t.client == nil {
		credential := auth.NewCredential(
			common.Env("TENCENT_ACCESS_KEY_ID"),
			common.Env("TENCENT_ACCESS_KEY_SECRET"),
		)

		cpf := profile.NewClientProfile()
		cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"

		client, err := lighthouse.NewClient(credential, t.GetRegion(), cpf)

		if err != nil {
			panic("create tencent lighthouse client error:" + err.Error())
		}

		t.client = client
	}

	return t.client
}

func (t *TencentLighthouse) GetInstances() ([]resource.InstanceInfo, error) {
	return t.getInstances()
}

func (t *TencentLighthouse) getInstances() ([]resource.InstanceInfo, error) {
	list := make([]resource.InstanceInfo, 0)

	offset := 0

	for {
		request := lighthouse.NewDescribeInstancesRequest()
		request.Limit = tea.Int64(100)
		request.Offset = tea.Int64(int64(offset))

		result, err := t.getClient().DescribeInstances(request)

		if err != nil {
			return nil, err
		}

		for _, item := range result.Response.InstanceSet {
			item := item

			ipAddress := ""

			if t.GetIPType() == "public" {
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
			offset += 100
			continue
		}

		break
	}

	return list, nil
}

func (t *TencentLighthouse) GetIPType() string {
	if common.Env("TENCENT_PUBLIC_IP") == "0" {
		return "private"
	}

	return "public"
}

func (t *TencentLighthouse) GetRegion() string {
	return common.Env("TENCENT_REGION")
}
