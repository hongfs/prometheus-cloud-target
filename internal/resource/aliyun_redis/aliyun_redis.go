package aliyun_redis

import (
	"errors"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	rkvstore20150101 "github.com/alibabacloud-go/r-kvstore-20150101/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/duke-git/lancet/v2/random"
	"github.com/hongfs/prometheus-cloud-target/internal/common"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"github.com/zeromicro/go-zero/core/threading"
	"log"
	"strconv"
)

type AliyunRedis struct {
	client *rkvstore20150101.Client
}

func (a *AliyunRedis) getClient() *rkvstore20150101.Client {
	if a.client == nil {
		config := &openapi.Config{
			AccessKeyId:     tea.String(common.Env("ALIYUN_ACCESS_KEY_ID")),
			AccessKeySecret: tea.String(common.Env("ALIYUN_ACCESS_KEY_SECRET")),
			RegionId:        tea.String(a.GetRegion()),
		}

		client, err := rkvstore20150101.NewClient(config)

		if a.GetIPType() == "private" {
			config.Network = tea.String("vpc")
		}

		if err != nil {
			panic("create aliyun redis client error:" + err.Error())
		}

		a.client = client
	}

	return a.client
}

func (a *AliyunRedis) GetInstances() ([]resource.InstanceInfo, error) {
	list, err := a.getInstances()

	if err != nil {
		return nil, err
	}

	wg := threading.NewRoutineGroup()

	ch := make(chan bool, 20)

	for i, item := range list {
		ch <- true

		i := i
		item := item

		wg.Run(func() {
			defer func() {
				<-ch
			}()

			info, err := a.getInstancesInfo(item)

			if err != nil {
				log.Println("getInstancesInfo error:", err.Error())
				return
			}

			list[i] = *info
		})
	}

	wg.Wait()

	return list, nil
}

func (a *AliyunRedis) getInstances() ([]resource.InstanceInfo, error) {
	list := make([]resource.InstanceInfo, 0)

	var page int32 = 1

	for {
		result, err := a.getClient().DescribeInstancesWithOptions(&rkvstore20150101.DescribeInstancesRequest{
			RegionId:       tea.String(a.GetRegion()),
			InstanceStatus: tea.String("Normal"),
			NetworkType:    tea.String("VPC"),
			PageNumber:     tea.Int32(page),
			PageSize:       tea.Int32(50),
			Expired:        tea.String("false"),
			GlobalInstance: tea.Bool(false),
		}, &util.RuntimeOptions{
			Autoretry:   tea.Bool(true),
			MaxAttempts: tea.Int(3),
		})

		if err != nil {
			return nil, err
		}

		for _, item := range result.Body.Instances.KVStoreInstance {
			list = append(list, resource.InstanceInfo{
				Type: resource.RedisInstanceType,
				ID:   *item.InstanceId,
			})
		}

		if len(result.Body.Instances.KVStoreInstance) == 50 {
			page++
			continue
		}

		break
	}

	return list, nil
}

func (a *AliyunRedis) getInstancesInfo(info resource.InstanceInfo) (*resource.InstanceInfo, error) {
	result, err := a.getClient().DescribeDBInstanceNetInfo(&rkvstore20150101.DescribeDBInstanceNetInfoRequest{
		InstanceId: tea.String(info.ID),
	})

	if err != nil {
		return nil, err
	}

	for _, network := range result.Body.NetInfoItems.InstanceNetInfo {
		if *network.IPType == "Public" && a.GetIPType() != "public" {
			continue
		} else if *network.IPType == "Private" && a.GetIPType() != "private" {
			continue
		}

		port, _ := strconv.Atoi(*network.Port)

		info.PublicAddress = *network.ConnectionString
		info.PublicPort = uint16(port)
	}

	if info.PublicAddress == "" {
		if a.GetIPType() == "private" {
			return nil, errors.New("private address is empty for instance " + info.ID)
		}

		connectionPrefix := info.ID + random.RandString(4)

		if len(connectionPrefix) > 40 {
			connectionPrefix = connectionPrefix[:40]
		} else if len(connectionPrefix) < 5 {
			connectionPrefix = random.RandString(30)
		}

		input := &rkvstore20150101.AllocateInstancePublicConnectionRequest{
			InstanceId:             tea.String(info.ID),
			ConnectionStringPrefix: tea.String(connectionPrefix),
			Port:                   tea.String(fmt.Sprintf("%d", random.RandInt(1000, 6000))),
		}

		_, err = a.getClient().AllocateInstancePublicConnection(input)

		if err != nil {
			return nil, err
		}

		port, err := strconv.Atoi(*input.Port)

		if err != nil {
			return nil, err
		}

		info.PublicAddress = fmt.Sprintf("%s.redis.rds.aliyuncs.com", connectionPrefix)
		info.PublicPort = uint16(port)
	}

	if info.PublicAddress == "" {
		return nil, errors.New("public address is empty for instance " + info.ID)
	}

	accounts, err := a.getClient().DescribeAccounts(&rkvstore20150101.DescribeAccountsRequest{
		InstanceId:  tea.String(info.ID),
		AccountName: tea.String(a.GetUsername()),
	})

	if err != nil {
		return nil, err
	}

	if len(accounts.Body.Accounts.Account) == 0 {
		_, err = a.getClient().CreateAccount(&rkvstore20150101.CreateAccountRequest{
			InstanceId:         tea.String(info.ID),
			AccountName:        tea.String(a.GetUsername()),
			AccountPassword:    tea.String(a.GetPassword()),
			AccountPrivilege:   tea.String("RoleReadOnly"),
			AccountDescription: tea.String("Prometheus"),
			AccountType:        tea.String("Normal"),
		})

		if err != nil {
			return nil, err
		}
	} else {
		_, err = a.getClient().ResetAccountPassword(&rkvstore20150101.ResetAccountPasswordRequest{
			InstanceId:      tea.String(info.ID),
			AccountName:     tea.String(a.GetUsername()),
			AccountPassword: tea.String(a.GetPassword()),
		})

		if err != nil {
			return nil, err
		}
	}

	info.Username = a.GetUsername()
	info.Password = a.GetPassword()

	return &info, nil
}

func (a *AliyunRedis) GetUsername() string {
	return common.Env("REDIS_USERNAME")
}

func (a *AliyunRedis) GetPassword() string {
	return common.Env("REDIS_PASSWORD")
}

func (a *AliyunRedis) GetIPType() string {
	if common.Env("ALIYUN_PUBLIC_IP") == "0" {
		return "private"
	}

	return "public"
}

func (a *AliyunRedis) GetRegion() string {
	return common.Env("ALIYUN_REGION")
}
