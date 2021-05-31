package ginHandler

import (
  "github.com/gin-contrib/cors"
)

func CorsConfig() cors.Config {
  corsConf := cors.DefaultConfig()
  corsConf.AllowOrigins = []string{"http://127.0.0.1:8080","http://127.0.0.1:8000"}
  corsConf.AllowMethods = []string{"GET", "POST"}
  corsConf.AllowHeaders = []string{"Authorization", "Content-Type", "Upgrade", "Origin",
    "Connection", "Accept-Encoding", "Accept-Language", "Host", "Access-Control-Request-Method",
    "Access-Control-Request-Headers"}
  return corsConf
}
