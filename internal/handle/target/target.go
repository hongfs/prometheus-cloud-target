package target

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hongfs/prometheus-cloud-target/internal/handle/instance"
	"github.com/hongfs/prometheus-cloud-target/internal/resource"
	"net/http"
)

type Group struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func Handle(c *gin.Context) {
	groups, err := handle(c)

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, groups)
}

func handle(c *gin.Context) ([]Group, error) {
	instancesType, ok := map[string]resource.InstanceType{
		"ecs":   resource.EcsInstanceType,
		"mysql": resource.MySQLInstanceType,
		"redis": resource.RedisInstanceType,
	}[c.Query("type")]

	if !ok {
		return nil, errors.New("invalid type")
	}

	instances := instance.Instances

	var groups = make([]Group, 0)

	for _, item := range instances {
		if item.Type != instancesType {
			continue
		}

		if item.PublicAddress == "" {
			continue
		}

		address := fmt.Sprintf("%s:%d", item.PublicAddress, item.PublicPort)

		if item.Type == resource.RedisInstanceType {
			address = "redis://" + address
		}

		groups = append(groups, Group{
			Targets: []string{address},
			Labels: map[string]string{
				"ID": item.ID,
			},
		})
	}

	return groups, nil
}
