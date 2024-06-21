# Sentry Prometheus Exporter developed in the Golang language 👋

[![License: GNU General Public License v2.0](https://img.shields.io/github/license/italux/sentry-prometheus-exporter-go)](https://github.com/italux/sentry-prometheus-exporter-go/blob/master/LICENSE)
![Dockerhub Build](https://img.shields.io/docker/cloud/automated/italux/sentry-prometheus-exporter-go)
![Dockerhub build status](https://img.shields.io/docker/cloud/build/italux/sentry-prometheus-exporter-go)
![Version](https://img.shields.io/github/v/tag/italux/sentry-prometheus-exporter-go)

> Export the Sentry data to the measurements of the Sentry project in the same format as the Prometheus specification !!! 🚀


## ✈️ 开始

### 条件

* Golang >= 1.20.10
* Sentry API [auth token](https://docs.sentry.io/api/auth/#auth-tokens)
    > 身份验证令牌权限: `project:read` `org:read` `project:releases` `event:read`

### 安装

```sh
go mod tidy
```

### 启动

**创建 `.env` 文件**
```sh
export SENTRY_API_BASE_URL="https://sentry.io/api/0/"
export SENTRY_AUTH_TOKEN="[REPLACE_TOKEN]"
export SENTRY_EXPORTER_ORG_SLUG="[organization_slug]"
export SENTRY_EXPORTER_PROJECTS=""
export SENTRY_ISSUES_1H="True"
export SENTRY_ISSUES_24H="False"
export SENTRY_ISSUES_14D="False"
export SENTRY_RATE_LIMIT_METRICS="True"
export SENTRY_ISSUE_METRICS="True"
export SENTRY_EVENTS_METRICS="True"
export EXPORTER_PORT="8080"

```
```sh
source .env
```
```sh
go run main.go
```

## 📈 指标
* `sentry_open_issue_events`: A Number of open issues (aka is:unresolved) per project in the past 1h
* `sentry_open_issues_histogram`: Gauge Histogram of open issues split into 3 buckets: 1h, 24h, and 14d
* `sentry_events`: Total events counts per project
* `sentry_rate_limit_events_sec`: Rate limit of errors per second accepted for a project.

### Sentry Project 配置

- 默认情况下，将轮询sentry的API以检索所有项目。如果您希望对特定项目进行刮除，您可以执行以下操作
```sh
export SENTRY_EXPORTER_PROJECTS="project1,project2,project3"
```

### 指标配置

- 除了rate-limit-events指标外，默认情况下所有指标都被抓取，但是，可以通过将相关变量设置为False来禁用问题或事件相关指标；
```sh
export SENTRY_SCRAPE_ISSUE_METRICS=False
export SENTRY_SCRAPE_EVENT_METRICS=False
```

- 通过将相关变量设置为True来启用rate-limit-events度量；
```sh
export SENTRY_SCRAPE_RATE_LIMIT_METRICS=True
```

- 默认情况下，如果“SENTRY_SCRAPE_ISSUE_METRICS=True或未设置”，则抓取“1小时”，“24小时”和“14天”的问题指标。这些都可以通过将相关变量设置为False来禁用；
```sh
export SENTRY_ISSUES_1H=False
export SENTRY_ISSUES_24H=False
export SENTRY_ISSUES_14D=False
```

- ServiceMonitor 配置参考
```yaml
scrape_configs:
  - job_name: 'sentry_exporter'
    static_configs:
    - targets: ['sentry-exporter:9790']
    scrape_interval: 5m
    scrape_timeout: 4m
```

## 🏷 提醒建议

- 最小使用 scrape_interval: 5m。
- 这个值将由新问题和事件的数量来定义 更多的事件将需要更多的时间
- 对于导出器任务，使用高 scrape_timeout
- 一般建议设置为 scrape_interval - 1 (例如：4m)
- 如果禁用了特定指标的抓取，根据您的设置，上述值可以减少。

## 📝 License

Copyright © 2024 [Dz6666](https://github.com/Dz6666).
This project is [GNU General Public License v2.0](https://github.com/italux/sentry-prometheus-exporter/blob/master/LICENSE) licensed.

## 📒 文档

[Sentry Prometheus Exporter documentation](https://github.com/Dz6666/sentry-prometheus-exporter-go/)

## ⭐️ 作者

👤 **Daizhe**

* Github: [@Dz6666](https://github.com/Dz6666)
* Gitee: [@Dz6666](https://gitee.com/dz6666)
