package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"jinadam.github.io/t4spider/common"
	"jinadam.github.io/t4spider/csp"
)

var Port int

func main() {
	flag.IntVar(&Port, "port", 53001, `program port`)
	gin.SetMode("debug")
	router := NewRouter()
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(Port),
		Handler:        router,
		ReadTimeout:    1 * time.Duration(time.Second),
		WriteTimeout:   3 * time.Duration(time.Second),
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatal("startup service failed...", zap.Error(err))
	}
}

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.RecoveryWithWriter(gin.DefaultErrorWriter))
	// 默认
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"data": "pong",
		})
	})
	// 直播
	l := r.Group("/csp")
	{
		l.GET("/:spiderName", SpiderRequest)
	}
	return r
}

func SpiderRequest(c *gin.Context) {
	spiderName := c.Param("spiderName")
	var spider common.Spider
	switch spiderName {
	case "jianpian":
		{
			spider = new(csp.JianPianSpider)
		}
	default:
		{
			c.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}
	}

	if ids := c.Query("ids"); ids != "" {
		dr := spider.Detail(ids)
		c.PureJSON(http.StatusOK, dr)
	} else if wd := c.Query("wd"); wd != "" {
		v := spider.Search(wd, false)
		c.PureJSON(http.StatusOK, v)
	} else if play := c.Query("play"); play != "" {
		p := spider.Play(play)
		c.PureJSON(http.StatusOK, p)
	} else if t := c.Query("t"); t != "" {
		pg := c.DefaultQuery("pg", "1")
		ext := c.DefaultQuery("ext", "e30=")
		ext, _ = url.QueryUnescape(ext)
		i, _ := strconv.Atoi(pg)
		p := spider.Cate(t, i, ext)
		c.PureJSON(http.StatusOK, p)
	} else {
		p := spider.Home()
		c.PureJSON(http.StatusOK, p)
	}
}
