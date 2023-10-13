package instance

import (
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"github.com/hongfs/prometheus-cloud-target/internal/resource/aliyun_ecs"
	"github.com/hongfs/prometheus-cloud-target/internal/resource/aliyun_mysql"
	"github.com/hongfs/prometheus-cloud-target/internal/resource/aliyun_redis"
	"github.com/zeromicro/go-zero/core/threading"
	"log"
	"sync"
)

var Instances = make([]resource.InstanceInfo, 0)

func Load() ([]resource.InstanceInfo, error) {
	list := make([]resource.InstanceInfo, 0)

	mu := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	services := []resource.Cloud{
		&aliyun_ecs.AliyunEcs{},
		&aliyun_mysql.AliyunMySQL{},
		&aliyun_redis.AliyunRedis{},
	}

	for _, service := range services {
		service := service

		wg.Add(1)

		threading.GoSafe(func() {
			defer wg.Done()

			instances, err := service.GetInstances()

			if err != nil {
				log.Println("getInstances error:", err.Error())
				return
			}

			mu.Lock()
			list = append(list, instances...)
			mu.Unlock()
		})
	}

	wg.Wait()

	return list, nil
}
