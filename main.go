package main

import (
	"context"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/wujunwei928/parse-video/parser"

	"github.com/gin-gonic/gin"
)

type HttpResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func main() {
	r := gin.Default()
	k := cache.New(time.Hour, 2*time.Hour)

	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "framework7.html", gin.H{
			"title": os.Getenv("APP_NAME"),
		})
	})

	r.GET("/video/share/url/parse", func(c *gin.Context) {
		urlReg := regexp.MustCompile(`http[s]?:\/\/[\w.-]+[\w\/-]*[\w.-]*\??[\w=&:\-\+\%]*[/]*`)
		videoShareUrl := urlReg.FindString(c.Query("url"))
		//fmt.Println("123", videoShareUrl, c.Query("url"))

		// 从缓存中获取
		if res, ok := k.Get(videoShareUrl); ok {
			c.JSON(http.StatusOK, res)
			return
		}

		parseRes, err := parser.ParseVideoShareUrl(videoShareUrl)
		jsonRes := HttpResponse{
			Code: 200,
			Msg:  "解析成功",
			Data: parseRes,
		}
		if err != nil {
			jsonRes = HttpResponse{
				Code: 201,
				Msg:  "解析失败",
			}
		}

		c.JSON(http.StatusOK, jsonRes)
		k.Set(videoShareUrl, jsonRes, cache.DefaultExpiration)
	})

	r.GET("/video/id/parse", func(c *gin.Context) {
		videoId := c.Query("video_id")
		source := c.Query("source")

		parseRes, err := parser.ParseVideoId(videoId, source)
		jsonRes := HttpResponse{
			Code: 200,
			Msg:  "解析成功",
			Data: parseRes,
		}
		if err != nil {
			jsonRes = HttpResponse{
				Code: 201,
				Msg:  "解析失败",
			}
		}

		c.JSON(200, jsonRes)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器 (设置 5 秒的超时时间)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")

}
