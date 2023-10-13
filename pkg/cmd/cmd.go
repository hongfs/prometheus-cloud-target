package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/hongfs/prometheus-cloud-target/internal/handle/reload"
	"github.com/hongfs/prometheus-cloud-target/internal/handle/target"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"github.com/zeromicro/go-zero/core/threading"
	"time"
)

var Instances = make([]resource.InstanceInfo, 0)

func Start(addr string) error {
	threading.GoSafe(async)

	r := gin.Default()

	r.GET("/target", target.Handle)
	r.GET("/-/reload", reload.Handle)

	return r.Run(addr)
}

func async() {
	for {
		reload.Handle(nil)

		time.Sleep(time.Second * 30)
	}
}
