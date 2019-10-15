package main

import (
	"github.com/gin-gonic/gin"
	. "wechatrobot/tools"
)

func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.GET("/wechat/talk", CheckToken)

	r.GET("/wechat/getAccessToken", func(c *gin.Context) {
		c.String(200, GetAccessToken())
	})

	r.POST("/wechat/talk", RobotResponse)

	r.GET("/toutiaoFocus", func(c *gin.Context) {
		c.JSON(200, TouTiaoFocus())
	})

	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}
