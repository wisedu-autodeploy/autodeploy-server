package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/satori/go.uuid"

	"gopkg.in/olahol/melody.v1"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/gin/json"
	"github.com/lisiur/autodeploy/gitlab"
	"github.com/lisiur/autodeploy/marathon"
)

var commands = map[string]string{
	"windows": "start",
	"darwin":  "open",
	"linux":   "xdg-open",
}

var deploySessions = map[string]*DeploySession{}
var websocketSessions = map[string]*melody.Session{}

// DeploySession .
// Status 0 -> init, -1 -> failed, 1 -> running, 2 -> succeed
type DeploySession struct {
	Params DeployCfg `json:"params,omitempty"`
	Step   int       `json:"step,omitempty"`
	Log    []string  `json:"log,omitempty"`
	Status int       `json:"status,omitempty"`
	Tag    string    `json:"tag,omitempty"`
	Image  string    `json:"image,omitempty"`
}

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
	MarathonID   string `json:"marathon_id,omitempty"`
}

// Open calls the OS default program for uri
func Open(uri string) error {
	run, ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}

	cmd := exec.Command(run, uri)
	return cmd.Start()
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	m := melody.New()

	router.Use(cors.Default())
	// router.StaticFS("static", http.Dir("web/static"))
	router.StaticFS("static", assetFS())

	router.GET("/", func(c *gin.Context) {
		router.LoadHTMLFiles("web/app.html")
		c.HTML(http.StatusOK, "app.html", gin.H{
			"title": "Welcome",
		})
	})

	router.GET("/deploy/:uuid", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})
	m.HandleConnect(func(s *melody.Session) {
		//s.Write([]byte("test"))
	})
	m.HandleMessage(func(s *melody.Session, msg []byte) {
		UUID := string(msg)
		deploySession, ok := deploySessions[UUID]
		websocketSessions[UUID] = s
		if !ok {
			res := map[string]interface{}{
				"code":    "-1",
				"message": "not found UUID",
				"data":    nil,
			}
			jsonRes, _ := json.Marshal((res))
			s.Write(jsonRes)
		} else {
			if deploySession.Status == 0 { // 未开始
				go deploy(deploySession, UUID)
			} else {
				res := map[string]interface{}{
					"code":    "0",
					"message": "",
					"data":    deploySession,
				}
				jsonRes, _ := json.Marshal((res))
				s.Write(jsonRes)
			}
		}
	})

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

		v1.POST("/autodeploy", func(c *gin.Context) {
			var config DeployCfg
			var UUID = uuid.Must(uuid.NewV4()).String()

			c.BindJSON(&config)

			deploySessions[UUID] = &DeploySession{
				Params: config,
				Step:   -1,
				Log:    []string{},
			}

			c.JSON(200, gin.H{
				"code":    "0",
				"message": "success",
				"data": map[string]interface{}{
					"uuid": UUID,
				},
			})
			return
		})
	}

	err := Open("http://localhost:2334/static/app.html")
	if err != nil {
		log.Println(err)
	}

	router.Run(":2334")
}

func deploy(ds *DeploySession, UUID string) {
	var err error
	var config = ds.Params
	var res interface{}
	var jsonRes []byte

	var s, ok = websocketSessions[UUID]
	if !ok {
		return
	}

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
		MarathonID:   config.MarathonID,
	}

	// tag
	ds.Step = 0
	ds.Status = 1 // running
	res = map[string]interface{}{
		"code":    "0",
		"message": "",
		"data":    ds,
	}
	jsonRes, _ = json.Marshal(res)

	s, ok = websocketSessions[UUID]
	if !ok {
		return
	}
	s.Write(jsonRes)

	_, err = gitlab.Init(gitlabCfg)
	if err != nil {
		return
	}

	log.Println("new tag")
	tag, err := gitlab.NewTag(marathonCfg)
	if err != nil {
		return
	}

	// building
	ds.Step = 1
	ds.Tag = tag
	res = map[string]interface{}{
		"code":    "0",
		"message": "",
		"data":    ds,
	}
	jsonRes, _ = json.Marshal(res)

	s, ok = websocketSessions[UUID]
	if !ok {
		return
	}
	s.Write(jsonRes)

	log.Println("building")
	ok, _, image, err := gitlab.WatchBuildLog(marathonCfg, tag, false)
	if err != nil || !ok {
		return
	}

	// deploy
	ds.Step = 2
	ds.Image = image
	res = map[string]interface{}{
		"code":    "0",
		"message": "",
		"data":    ds,
	}
	jsonRes, _ = json.Marshal(res)

	s, ok = websocketSessions[UUID]
	if !ok {
		return
	}
	s.Write(jsonRes)

	log.Println("deploying")
	ok, err = marathon.Deploy(marathonCfg, image)
	if err != nil || !ok {
		log.Println(err)
		res = map[string]interface{}{
			"code":    "-1",
			"message": err.Error(),
			"data":    ds,
		}
		jsonRes, _ = json.Marshal(res)

		s, ok = websocketSessions[UUID]
		if !ok {
			return
		}
		s.Write(jsonRes)

		return
	}

	ds.Step = 3
	ds.Status = 2 // succeed
	res = map[string]interface{}{
		"code":    "0",
		"message": "",
		"data":    ds,
	}
	jsonRes, _ = json.Marshal(res)

	s, ok = websocketSessions[UUID]
	if !ok {
		return
	}
	s.Write(jsonRes)

	log.Println("done")
	return
}
