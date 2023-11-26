package instance

import (
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"github.com/hongfs/prometheus-cloud-target/internal/resource/aliyun_ecs"
	"github.com/hongfs/prometheus-cloud-target/internal/resource/aliyun_mysql"
	"github.com/hongfs/prometheus-cloud-target/internal/resource/aliyun_redis"
	"github.com/hongfs/prometheus-cloud-target/internal/resource/aliyun_swas"
	"github.com/hongfs/prometheus-cloud-target/internal/resource/tencent_lighthouse"
	"github.com/zeromicro/go-zero/core/threading"
	"log"
	"os"
	"sync"
)

var Instances = make([]resource.InstanceInfo, 0)

func Load() ([]resource.InstanceInfo, error) {
	list := make([]resource.InstanceInfo, 0)

	services := []resource.Cloud{
		&aliyun_ecs.AliyunEcs{},
		&aliyun_mysql.AliyunMySQL{},
		&aliyun_redis.AliyunRedis{},
		&aliyun_swas.AliyunSwas{},
	}

	if os.Getenv("TENCENT_ENABLE") == "1" {
		services = append(
			services,
			&tencent_lighthouse.TencentLighthouse{},
		)
	}

	mu := new(sync.Mutex)
	wg := threading.NewRoutineGroup()

	for _, service := range services {
		service := service

		wg.Run(func() {
			instances, err := service.GetInstances()

			if err != nil {
				log.Println("getInstances error:", err.Error())
				return
			}

			mu.Lock()
			defer mu.Unlock()
			list = append(list, instances...)
		})
	}

	wg.Wait()

	return list, nil
}
