package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"gopkg.in/routeros.v2"
	"gopkg.in/routeros.v2/proto"
	"log"
	"strings"
	"time"
)

type Device struct {
	Address  string `form:"Address"`
	Username string `form:"Username"`
	Password string `form:"Password"`
	Ssl      string
	PortAPI  uint16 `form:"PortApi"`
	Command  string `form:"cmd"`
	Args     string `form:"сargs"`
	Debug    string `form:"debug"`
	Format   string `form:"format"`
	Ot       []*proto.Sentence
}

var Dev Device

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	tbox := rice.MustFindBox("templates")
	render := multitemplate.New()
	list := [...]string{"home.html"}
	for _, x := range list {
		templateString, err := tbox.String(x)
		if err != nil {
			fmt.Println(err)
		}
		render.AddFromString(x, templateString)
	}
	r.HTMLRender = render
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "home.html", gin.H{})
	})
	r.StaticFS("/static", rice.MustFindBox("templates").HTTPBox())
	r.POST("/send", send)
	r.Run("127.0.0.1:8081")
}

func send(c *gin.Context) {
	err := c.Bind(&Dev)
	if c.PostForm("ssl") != "" {
		Dev.Ssl = "1"
	} else {
		Dev.Ssl = "0"
	}
	if err != nil {
		c.JSON(200, erJson(err.Error()))
		c.Done()
		return
	}
	Devcon, err := dial()
	if err != nil {
		log.Println(Dev)
		log.Println(err)
		c.JSON(200, erJson(err.Error()))
		//Devcon.Close()
		c.Done()
		return
	}
	ar := strings.Split(Dev.Command+"\n"+Dev.Args, "\n")
	log.Println(ar)
	r, err := Devcon.RunArgs(ar)
	if err != nil {
		c.JSON(200, erJson(err.Error()))
		Devcon.Close()
		c.Done()
		return

	}
	Dev.Ot = r.Re
	Devcon.Close()
	c.JSON(200, rJson())
}

func dial() (*routeros.Client, error) {
	var Addr = fmt.Sprintf("%s:%d", Dev.Address, Dev.PortAPI)
	if Dev.Ssl == "1" {
		return routeros.DialTLS(Addr, Dev.Username, Dev.Password, &tls.Config{InsecureSkipVerify: true})
	}
	timeout := time.Duration(4)
	return routeros.DialTimeout(Addr, Dev.Username, Dev.Password, timeout*time.Second)
}
func erJson(e string) (s string) {
	var myerror []string
	myerror = append(myerror, e)
	r, _ := json.Marshal(myerror)
	return string(r)
}
func rJson() (s string) {
	var otv = make(map[int]map[string]string)
	if Dev.Debug == "1" {

		otvjs, _ := json.Marshal(Dev.Ot)
		return string(otvjs)
	}

	for i, k := range Dev.Ot {
		otv[i] = k.Map
	}
	otvjs, _ := json.Marshal(otv)
	return string(otvjs)
}
