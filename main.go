package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"sync"

	"github.com/satori/go.uuid"

	"gopkg.in/olahol/melody.v1"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/gin/json"
	"github.com/wisedu-autodeploy/autodeploy/gitlab"
	"github.com/wisedu-autodeploy/autodeploy/marathon"
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

// BaseResponse .
type BaseResponse struct {
	Code    string      `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Open calls the OS default program for uri
func Open(uri string) error {
	run, ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}

	if runtime.GOOS == "windows" {
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", uri).Start()
	}
	return exec.Command(run, uri).Start()
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	m := melody.New()

	router.Use(cors.Default())
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
		if !ok { // not found related deploy session
			jsonRes := jsonResponse("-1", nil, "not found UUID")
			s.Write(jsonRes)
		} else { // found related deploy session
			if deploySession.Status == 0 { // 未开始
				go deploy(deploySession, UUID)
			} else {
				jsonRes := jsonResponse("0", deploySession, "")
				s.Write(jsonRes)
			}
		}
	})

	v1 := router.Group("/v1/api")
	{
		v1.POST("/login", func(c *gin.Context) {
			var userInfo UserInfo
			c.BindJSON(&userInfo)
			gitlabUser := gitlab.User{
				Username: userInfo.Username,
				Password: userInfo.Password,
			}
			gitlabApps, err := gitlab.GetAllApps(gitlabUser)
			if err != nil {
				handleErr(c, err)
				return
			}

			marathonApps, err := marathon.GetApps()
			if err != nil {
				handleErr(c, err)
				return
			}

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
	// gitlab.Debugger()
	var err error
	var config = ds.Params
	var res interface{}
	var jsonRes []byte

	var s, ok = websocketSessions[UUID]
	if !ok {
		return
	}

	var gitlabParams = gitlab.Params{
		User: gitlab.User{
			Username: config.Username,
			Password: config.Password,
		},
		Project: gitlab.Project{
			Maintainer: config.Maintainer,
			Name:       config.Name,
		},
	}
	var marathonCfg = marathon.Config{
		MarathonName: config.MarathonName,
		MarathonID:   config.MarathonID,
	}

	// tag
	ds.Step = 0
	ds.Status = 1 // running
	res = BaseResponse{
		Code:    "0",
		Data:    ds,
		Message: "",
	}
	putMsg(res, UUID)

	tag, err := gitlab.NewTag(gitlabParams)
	if err != nil {
		return
	}

	// building
	ds.Step = 1
	ds.Tag = tag
	res = BaseResponse{
		Code:    "0",
		Data:    ds,
		Message: "",
	}
	putMsg(res, UUID)

	var wg sync.WaitGroup
	logChan := make(chan *gitlab.Logger)
	wg.Add(2)

	go gitlab.WatchBuildLog(gitlabParams, tag, logChan, &wg) // 写日志
	go gitlab.GetBuildLog(func(logger *gitlab.Logger) {
		ds.Log = logger.Log
		res = BaseResponse{
			Code:    "0",
			Data:    ds,
			Message: "",
		}
		putMsg(res, UUID)
	}, logChan, &wg)

	wg.Wait()

	// deploy
	logger := <-logChan
	close(logChan)
	image := logger.Image
	ds.Step = 2
	ds.Image = image
	res = BaseResponse{
		Code:    "0",
		Data:    ds,
		Message: "",
	}
	putMsg(res, UUID)

	ok, err = marathon.Deploy(marathonCfg, image)
	if err != nil || !ok {
		log.Println(err)
		jsonRes = jsonResponse("-1", ds, err.Error())

		s, ok = websocketSessions[UUID]
		if !ok {
			return
		}
		s.Write(jsonRes)

		return
	}

	ds.Step = 3
	ds.Status = 2 // succeed
	jsonRes = jsonResponse("0", ds, "")

	s, ok = websocketSessions[UUID]
	if !ok {
		return
	}
	s.Write(jsonRes)

	// TODO: delete websocketSessions[UUID]
	return
}

func jsonResponse(code string, data interface{}, message string) (jsonRes []byte) {
	res := BaseResponse{
		Code:    code,
		Data:    data,
		Message: message,
	}
	jsonRes, _ = json.Marshal(res)
	return
}

func handleErr(c *gin.Context, err error) {
	log.Println(err)
	c.JSON(200, gin.H{
		"code":    "-1",
		"message": err.Error(),
		"data":    nil,
	})
}

func putMsg(res interface{}, UUID string) {
	jsonRes, _ := json.Marshal(res)

	s, ok := websocketSessions[UUID]
	if !ok {
		return
	}
	s.Write(jsonRes)
}
