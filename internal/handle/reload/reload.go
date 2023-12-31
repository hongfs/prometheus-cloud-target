package reload

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/hongfs/prometheus-cloud-target/internal/handle/instance"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"github.com/zeromicro/go-zero/core/threading"
	"log"
	"net/http"
	"os"
	"sync"
)

var mu = new(sync.Mutex)

func Handle(c *gin.Context) {
	threading.GoSafe(func() {
		err := handle()

		if err != nil {
			log.Printf("handle error: %s", err.Error())
		}
	})

	if c == nil {
		return
	}

	c.Status(http.StatusOK)
}

func handle() error {
	mu.Lock()
	defer mu.Unlock()

	list, err := instance.Load()

	if err != nil {
		return err
	}

	instance.Instances = list

	data := make(map[string]interface{})

	for _, item := range list {
		if item.Type != resource.MySQLInstanceType {
			continue
		}

		if item.Username == "" || item.Password == "" {
			continue
		}

		data["client"] = map[string]string{
			"user":     item.Username,
			"password": item.Password,
		}

		break
	}

	if _, ok := data["client"]; !ok {
		return nil
	}

	content, err := gjson.Marshal(data)

	if err != nil {
		return err
	}

	json, err := gjson.LoadJson(content)

	if err != nil {
		return err
	}

	content, err = json.ToIni()

	if err != nil {
		return err
	}

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		configPath = "/data/my.cnf"
	}

	return os.WriteFile(configPath, content, os.ModePerm)
}
