package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"sentry-exporter/collector"
	"sentry-exporter/config"
	"sentry-exporter/sentry"
)

func main() {
	sentryAPI := sentry.NewSentryAPI(config.SentryAPIBaseURL, config.SentryAuthToken)
	resp, err := sentryAPI.Get("organizations/")
	if err != nil {
		log.Fatalf("Failed to get organizations: %v", err)
	}
	defer resp.Body.Close()

	//// 初始化 Sentry Collector
	//colle := collector.NewSentryCollector(sentryAPI, "sentry", []string{}, []bool{true, true, true, false, true, false})
	//// 获取 Sentry 数据测试
	//data := colle.FetchSentryData()
	//if data == nil {
	//	log.Fatalf("Failed to build Sentry data")
	//}
	//// 打印获取到的数据
	//fmt.Printf("Sentry Data:\n%v\n", data)

	// 初始化 Sentry Collector
	colle1 := collector.NewSentryCollector(sentryAPI, config.SentryExporterOrgSlug, []string{config.SentryExporterProjects},
		[]bool{config.SentryIssueMetrics,
			config.SentryEventsMetrics,
			config.SentryRateLimitMetrics,
			config.SentryIssues1H,
			config.SentryIssues24H,
			config.SentryIssues14D})

	// 注册收集器
	prometheus.MustRegister(colle1)
	router := gin.Default()
	// Home endpoint
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "<h1>Sentry Issues & Events Exporter</h1><h3>Go to <a href=/metrics/>/metrics</a></h3>")
	})
	// Healthz endpoint
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	// Metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		h := promhttp.Handler()
		h.ServeHTTP(c.Writer, c.Request)
	})
	// 启动 HTTP 服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s\n", port)
	log.Fatal(router.Run("0.0.0.0:" + config.EXPORTER_PORT))
}
