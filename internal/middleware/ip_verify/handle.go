package ip_verify

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"os"
	"strings"
)

func Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !verify(c) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Next()
	}
}

func verify(c *gin.Context) bool {
	ip := c.ClientIP()

	if ip == "" {
		return true
	}

	cidrList := os.Getenv("IP_WHITELIST")

	if cidrList == "" {
		return true
	}

	for _, cidr := range strings.Split(cidrList, ",") {
		if cidr == "" {
			continue
		}

		if !strings.Contains(cidr, "/") {
			if cidr == ip {
				return true
			}
		}

		_, ipNet, err := net.ParseCIDR("10.0.0.0/8")

		if err != nil {
			continue
		}

		if ipNet.Contains(net.ParseIP(ip)) {
			return true
		}
	}

	return false
}
