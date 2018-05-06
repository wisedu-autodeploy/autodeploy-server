package main

import (
	"log"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"

	"github.com/lisiur/autodeploy/gitlab"
	"github.com/lisiur/autodeploy/marathon"
)

// UserInfo .
type UserInfo struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// DeployCfg .
type DeployCfg struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Maintainer   string `json:"maintainer,omitempty"`
	Name         string `json:"name,omitempty"`
	MarathonName string `json:"marathon_name,omitempty"`
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	v1 := router.Group("/v1/api")
	{
		v1.POST("/login", func(c *gin.Context) {
			var userInfo UserInfo
			c.BindJSON(&userInfo)
			gitlabCfg := gitlab.Config{
				Origin:      "http://172.16.7.53:9090",
				LoginAction: "/users/sign_in",
				Username:    userInfo.Username,
				Password:    userInfo.Password,
			}
			_, err := gitlab.Init(gitlabCfg)
			if err != nil {
				c.JSON(200, gin.H{
					"code":    "-1",
					"message": err.Error(),
					"data":    nil,
				})
				return
			}
			gitlabApps, err := gitlab.GetAllApps()
			if err != nil {
				c.JSON(200, gin.H{
					"code":    "-1",
					"message": err.Error(),
					"data":    nil,
				})
				return
			}

			marathonApps := marathon.GetApps()

			c.JSON(200, gin.H{
				"code":    "0",
				"message": "success",
				"data": map[string]interface{}{
					"gitlabApps":   gitlabApps,
					"marathonApps": marathonApps,
				},
			})
			return
		})

		v1.POST("/autodeploy", deploy)
	}

	router.Run(":8080")
}

func deploy(c *gin.Context) {
	var err error
	var config DeployCfg
	c.BindJSON(&config)

	var gitlabCfg = gitlab.Config{
		Origin:      "http://172.16.7.53:9090",
		LoginAction: "/users/sign_in",
		Username:    config.Username,
		Password:    config.Password,
	}
	var marathonCfg = marathon.Config{
		Maintainer:   config.Maintainer,
		Name:         config.Name,
		MarathonName: config.MarathonName,
	}
	_, err = gitlab.Init(gitlabCfg)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    "-1",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	log.Println("new tag")
	tag, err := gitlab.NewTag(marathonCfg)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    "-1",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	log.Println("building")
	ok, _, image, err := gitlab.WatchBuildLog(marathonCfg, tag, true)
	if err != nil || !ok {
		c.JSON(200, gin.H{
			"code":    "-1",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	log.Println("deploying")
	ok, err = marathon.Deploy(marathonCfg.MarathonName, image)
	if err != nil || !ok {
		c.JSON(200, gin.H{
			"code":    "-1",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	if ok {
		log.Println("done")
		c.JSON(200, gin.H{
			"code":    "-1",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
}
