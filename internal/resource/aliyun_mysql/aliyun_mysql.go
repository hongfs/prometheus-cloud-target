package aliyun_mysql

import (
	"errors"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	rds20140815 "github.com/alibabacloud-go/rds-20140815/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/duke-git/lancet/v2/random"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"github.com/zeromicro/go-zero/core/threading"
	"log"
	"os"
	"strconv"
)

type AliyunMySQL struct {
	client *rds20140815.Client
}

func (a *AliyunMySQL) getClient() *rds20140815.Client {
	if a.client == nil {
		config := &openapi.Config{
			AccessKeyId:     tea.String(os.Getenv("ALIYUN_ACCESS_KEY_ID")),
			AccessKeySecret: tea.String(os.Getenv("ALIYUN_ACCESS_KEY_SECRET")),
			RegionId:        tea.String(a.GetRegion()),
		}

		client, err := rds20140815.NewClient(config)

		if a.GetIPType() == "private" {
			config.Network = tea.String("vpc")
		}

		if err != nil {
			panic("init client error:" + err.Error())
		}

		a.client = client
	}

	return a.client
}

func (a *AliyunMySQL) GetInstances() ([]resource.InstanceInfo, error) {
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

func (a *AliyunMySQL) getInstances() ([]resource.InstanceInfo, error) {
	list := make([]resource.InstanceInfo, 0)

	var next *string

	for {
		result, err := a.getClient().DescribeDBInstances(&rds20140815.DescribeDBInstancesRequest{
			RegionId:         tea.String(a.GetRegion()),
			Engine:           tea.String("MySQL"),
			DBInstanceStatus: tea.String("Running"),
			DBInstanceType:   tea.String("Primary"),
			PageSize:         tea.Int32(100),
			NextToken:        next,
		})

		if err != nil {
			return nil, err
		}

		for _, item := range result.Body.Items.DBInstance {
			list = append(list, resource.InstanceInfo{
				Type: resource.MySQLInstanceType,
				ID:   *item.DBInstanceId,
			})
		}

		if len(result.Body.Items.DBInstance) == 100 {
			next = result.Body.NextToken
			continue
		}

		break
	}

	return list, nil
}

func (a *AliyunMySQL) getInstancesInfo(info resource.InstanceInfo) (*resource.InstanceInfo, error) {
	result, err := a.getClient().DescribeDBInstanceNetInfo(&rds20140815.DescribeDBInstanceNetInfoRequest{
		DBInstanceId:             tea.String(info.ID),
		DBInstanceNetRWSplitType: tea.String("Normal"),
	})

	if err != nil {
		return nil, err
	}

	for _, network := range result.Body.DBInstanceNetInfos.DBInstanceNetInfo {
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

		input := &rds20140815.AllocateInstancePublicConnectionRequest{
			DBInstanceId:           tea.String(info.ID),
			ConnectionStringPrefix: tea.String(connectionPrefix),
			Port:                   tea.String(fmt.Sprintf("%d", random.RandInt(1000, 6000))),
		}

		result, err := a.getClient().AllocateInstancePublicConnection(input)

		if err != nil {
			return nil, err
		}

		port, err := strconv.Atoi(*input.Port)

		if err != nil {
			return nil, err
		}

		info.PublicAddress = *result.Body.ConnectionString
		info.PublicPort = uint16(port)
	}

	if info.PublicAddress == "" {
		return nil, errors.New("public address is empty for instance " + info.ID)
	}

	accounts, err := a.getClient().DescribeAccounts(&rds20140815.DescribeAccountsRequest{
		DBInstanceId: tea.String(info.ID),
		AccountName:  tea.String(a.GetUsername()),
	})

	if err != nil {
		return nil, err
	}

	if len(accounts.Body.Accounts.DBInstanceAccount) == 0 {
		_, err = a.getClient().CreateAccount(&rds20140815.CreateAccountRequest{
			DBInstanceId:       tea.String(info.ID),
			AccountName:        tea.String(a.GetUsername()),
			AccountPassword:    tea.String(a.GetPassword()),
			AccountDescription: tea.String("Prometheus"),
		})

		if err != nil {
			return nil, err
		}
	} else {
		_, err = a.getClient().ResetAccountPassword(&rds20140815.ResetAccountPasswordRequest{
			DBInstanceId:    tea.String(info.ID),
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

func (a *AliyunMySQL) GetUsername() string {
	return os.Getenv("MYSQL_USERNAME")
}

func (a *AliyunMySQL) GetPassword() string {
	return os.Getenv("MYSQL_PASSWORD")
}

func (a *AliyunMySQL) GetIPType() string {
	if os.Getenv("ALIYUN_PUBLIC_IP") == "0" {
		return "private"
	}

	return "public"
}

func (a *AliyunMySQL) GetRegion() string {
	return os.Getenv("ALIYUN_REGION")
}
