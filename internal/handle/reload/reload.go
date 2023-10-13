package reload

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/hongfs/prometheus-cloud-target/internal/handle/instance"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"log"
	"net/http"
	"os"
	"sync"
)

var mu = new(sync.Mutex)

func Handle(c *gin.Context) {
	err := handle(c)

	if err != nil {
		log.Printf("handle error: %s", err.Error())
	}

	c.Status(http.StatusOK)
}

func handle(c *gin.Context) error {
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
		return errors.New("no instance")
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
