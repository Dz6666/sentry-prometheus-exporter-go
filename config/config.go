package config

import (
	"log"
	"os"
	"strconv"
)

var (
	SentryAPIBaseURL       string
	SentryAuthToken        string
	SentryExporterOrgSlug  string
	SentryExporterProjects string
	SentryRateLimitMetrics bool
	SentryIssueMetrics     bool
	SentryEventsMetrics    bool
	SentryIssues1H         bool
	SentryIssues24H        bool
	SentryIssues14D        bool
	EXPORTER_PORT          string
)

func init() {
	// 从环境变量中读取配置
	SentryAPIBaseURL = os.Getenv("SENTRY_API_BASE_URL")
	SentryAuthToken = os.Getenv("SENTRY_AUTH_TOKEN")
	SentryExporterOrgSlug = os.Getenv("SENTRY_EXPORTER_ORG_SLUG")
	SentryExporterProjects = os.Getenv("SENTRY_EXPORTER_PROJECTS")
	SentryRateLimitMetrics, _ = strconv.ParseBool(os.Getenv("SENTRY_RATE_LIMIT_METRICS"))
	SentryIssueMetrics, _ = strconv.ParseBool(os.Getenv("SENTRY_ISSUE_METRICS"))
	SentryEventsMetrics, _ = strconv.ParseBool(os.Getenv("SENTRY_EVENTS_METRICS"))
	SentryIssues1H, _ = strconv.ParseBool(os.Getenv("SENTRY_ISSUES_1H"))
	SentryIssues24H, _ = strconv.ParseBool(os.Getenv("SENTRY_ISSUES_24H"))
	SentryIssues14D, _ = strconv.ParseBool(os.Getenv("SENTRY_ISSUES_14D"))
	EXPORTER_PORT = os.Getenv("EXPORTER_PORT")

	if SentryAPIBaseURL == "" {
		log.Printf("Error: SENTRY_API_BASE_URL environment variable is not set, defaulting to empty string.")
	}
	if SentryAuthToken == "" {
		log.Fatalf("Error: SENTRY_AUTH_TOKEN environment variable is not set.")
	}
	if SentryExporterOrgSlug == "" {
		log.Fatalf("Error: SENTRY_EXPORTER_ORG_SLUG environment variable is not set.")
	}
	if SentryExporterProjects == "" {
		log.Fatalf("Warning: SENTRY_EXPORTER_PROJECTS environment variable is not set.")
	}
	if !SentryRateLimitMetrics {
		SentryRateLimitMetrics = true
		log.Printf("Warning: SENTRY_RATE_LIMIT_METRICS is set to false. Rate limit metrics will not be collected.")
	}
	if !SentryIssueMetrics {
		SentryIssueMetrics = true
		log.Printf("Warning: SENTRY_ISSUE_METRICS is set to false. Issue metrics will not be collected.")
	}
	if !SentryEventsMetrics {
		SentryEventsMetrics = true
		log.Printf("Warning: SENTRY_EVENTS_METRICS is set to false. Events metrics will not be collected.")
	}
	if !SentryIssues1H && !SentryIssues24H && !SentryIssues14D {
		SentryIssues1H = true
		log.Printf("Warning: None of SENTRY_ISSUES_1H, SENTRY_ISSUES_24H, or SENTRY_ISSUES_14D is set to true. It's recommended to set at least one of them to true.")
	}
	if EXPORTER_PORT == "" {
		EXPORTER_PORT = "8080"
		log.Fatalf("Warning: EXPORTER_PORT environment variable is not set. Use the default 8080.")
	}
}
