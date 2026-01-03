package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/mrwestbury/empty-ns-manager/api/v1alpha1"
)

type WebServer struct {
	router   *gin.Engine
	policies *map[string]v1alpha1.EmptyNsPolicy
	version  string
}

func NewWebServer(policies *map[string]v1alpha1.EmptyNsPolicy, appVersion string) *WebServer {
	ws := &WebServer{
		version:  appVersion,
		policies: policies,
	}

	router := gin.Default()

	router.GET("/health", ws.HealthCheck)
	router.GET("/api/policies", ws.ListPolicies)
	ws.router = router
	return ws
}

func (ws *WebServer) Start() error {
	return ws.router.Run(":8080")
}

func (ws *WebServer) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"version": ws.version,
	})
}

func (ws *WebServer) ListPolicies(c *gin.Context) {
	c.JSON(200, ws.policies)
}
