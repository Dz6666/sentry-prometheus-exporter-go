package collector

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"sentry-exporter/sentry"
	"strconv"
	"time"
)

const (
	JSONCacheFile               = "./sentry-collector-exporter-cache.json"
	DefaultCacheExpireTimestamp = 2 * time.Minute
)

// SentryCollector 结构体
type SentryCollector struct {
	sentryAPI          *sentry.SentryAPI
	sentryOrgSlug      string
	sentryProjectsSlug []string
	issueMetrics       bool
	eventsMetrics      bool
	rateLimitMetrics   bool
	get1hMetrics       bool
	get24hMetrics      bool
	get14dMetrics      bool

	org          map[string]interface{}
	projectsData map[string]interface{}
}

// NewSentryCollector 函数用于创建 SentryCollector 实例
func NewSentryCollector(api *sentry.SentryAPI, orgSlug string, projectSlugs []string, metricConfig []bool) *SentryCollector {
	return &SentryCollector{
		sentryAPI:          api,
		sentryOrgSlug:      orgSlug,
		sentryProjectsSlug: projectSlugs,
		issueMetrics:       metricConfig[0],
		eventsMetrics:      metricConfig[1],
		rateLimitMetrics:   metricConfig[2],
		get1hMetrics:       metricConfig[3],
		get24hMetrics:      metricConfig[4],
		get14dMetrics:      metricConfig[5],
	}
}

// FetchSentryData is a public method to fetch Sentry data
func (c *SentryCollector) FetchSentryData() map[string]interface{} {
	return c.buildSentryDataFromAPI()
}

// buildSentryDataFromAPI 用于从 Sentry API 构建本地数据结构
func (c *SentryCollector) buildSentryDataFromAPI() map[string]interface{} {
	var data map[string]interface{}

	// 获取组织信息
	org, err := c.sentryAPI.GetOrg(c.sentryOrgSlug)
	if err != nil {
		log.Printf("Failed to fetch organization: %v\n", err)
		return nil
	}
	log.Printf("metadata: sentry organization: %s\n", org.Slug)

	// 初始化数据结构
	data = map[string]interface{}{
		"metadata": map[string]interface{}{
			"org":           org,
			"projects":      []interface{}{},
			"projects_slug": []string{},
			"projects_envs": map[string]interface{}{},
		},
	}

	//// TODO 打印 data 变量的值
	//dataJSON, err := json.MarshalIndent(data, "", "  ")
	//if err != nil {
	//	fmt.Println("Failed to marshal data:", err)
	//}
	//fmt.Println(string(dataJSON))

	// 如果指定了项目，则获取项目信息
	if len(c.sentryProjectsSlug) > 0 {
		log.Printf("metadata: projects specified: %d\n", len(c.sentryProjectsSlug))
		for _, projectSlug := range c.sentryProjectsSlug {
			log.Printf("metadata: getting %s project data from API\n", projectSlug)
			// 获取项目信息
			// dhgate dhgate-addressbook-service
			project, err := c.sentryAPI.GetProject(org.Slug, projectSlug)
			if err != nil {
				log.Printf("Failed to fetch project %s: %v\n", projectSlug, err)
				continue
			}
			//fmt.Println("project: ", project)

			// 更新data中的projects和projects_slug字段
			projects := data["metadata"].(map[string]interface{})["projects"].([]interface{})
			data["metadata"].(map[string]interface{})["projects"] = append(projects, project)
			projectsSlug := data["metadata"].(map[string]interface{})["projects_slug"].([]string)
			data["metadata"].(map[string]interface{})["projects_slug"] = append(projectsSlug, projectSlug)

			// 获取项目环境信息
			envs, err := c.sentryAPI.Environments(org.Slug, *project)
			if err != nil {
				log.Printf("Failed to fetch environments for project %s: %v\n", projectSlug, err)
				continue
			}
			// 更新data中的projects_envs字段
			projectsEnvs := data["metadata"].(map[string]interface{})["projects_envs"].(map[string]interface{})
			projectsEnvs[projectSlug] = envs

			// 构建项目问题数据
			if c.issueMetrics {
				projectsIssueData := make(map[string]map[string]map[string][]interface{})
				for _, env := range envs {
					projectsIssueData[projectSlug] = make(map[string]map[string][]interface{})

					if c.get1hMetrics {
						log.Printf("metadata: getting issues from API - project: %s env: %s age: 1h\n", projectSlug, env)
						issues1h, err := c.sentryAPI.Issues(org.Slug, *project, env, "1h")
						if err != nil {
							log.Printf("Failed to fetch issues for project %s, env %s, age 1h: %v\n", projectSlug, env, err)
							continue
						}
						// projectsIssueData[projectSlug][env]["1h"] = issues1h["all"].([]interface{})
						// 确保map的每一层都已初始化
						if _, ok := projectsIssueData[projectSlug]; !ok {
							projectsIssueData[projectSlug] = make(map[string]map[string][]interface{})
						}
						if _, ok := projectsIssueData[projectSlug][env]; !ok {
							projectsIssueData[projectSlug][env] = make(map[string][]interface{})
						}

						// 确保allIssues是一个切片，然后赋值
						allIssues, ok := issues1h["all"].([]interface{})
						if !ok {
							log.Printf("Unexpected type for 'all' key in issues data, expected []interface{} but got %T", issues1h["all"])
							continue
						}
						projectsIssueData[projectSlug][env]["1h"] = allIssues
					}

					if c.get24hMetrics {
						log.Printf("metadata: getting issues from API - project: %s env: %s age: 24h\n", projectSlug, env)
						issues24h, err := c.sentryAPI.Issues(org.Slug, *project, env, "24h")
						//fmt.Println(issues24h, err) // map[all:[] message:No issues found] <nil>

						if err != nil {
							log.Printf("Failed to fetch issues for project %s, env %s, age 24h: %v\n", projectSlug, env, err)
							continue
						}
						// projectsIssueData[projectSlug][env]["24h"] = issues24h["all"].([]interface{})
						// 确保map的每一层都已初始化
						if _, ok := projectsIssueData[projectSlug]; !ok {
							projectsIssueData[projectSlug] = make(map[string]map[string][]interface{})
						}

						if _, ok := projectsIssueData[projectSlug][env]; !ok {
							projectsIssueData[projectSlug][env] = make(map[string][]interface{})
						}

						// 确保allIssues是一个切片，然后赋值
						allIssues, ok := issues24h["all"].([]interface{})
						if !ok {
							log.Printf("Unexpected type for 'all' key in issues data, expected []interface{} but got %T", issues24h["all"])
							continue
						}
						projectsIssueData[projectSlug][env]["24h"] = allIssues
					}

					if c.get14dMetrics {
						log.Printf("metadata: getting issues from API - project: %s env: %s age: 14d\n", projectSlug, env)
						issues14d, err := c.sentryAPI.Issues(org.Slug, *project, env, "14d")
						if err != nil {
							log.Printf("Failed to fetch issues for project %s, env %s, age 14d: %v\n", projectSlug, env, err)
							continue
						}
						// projectsIssueData[projectSlug][env]["14d"] = issues14d["all"].([]interface{})
						// 确保map的每一层都已初始化
						if _, ok := projectsIssueData[projectSlug]; !ok {
							projectsIssueData[projectSlug] = make(map[string]map[string][]interface{})
						}
						if _, ok := projectsIssueData[projectSlug][env]; !ok {
							projectsIssueData[projectSlug][env] = make(map[string][]interface{})
						}

						// 确保allIssues是一个切片，然后赋值
						allIssues, ok := issues14d["all"].([]interface{})
						if !ok {
							log.Printf("Unexpected type for 'all' key in issues data, expected []interface{} but got %T", issues14d["all"])
							continue
						}
						projectsIssueData[projectSlug][env]["14d"] = allIssues
					}
				}
				data["projects_data"] = projectsIssueData
			}
		}
		log.Printf("metadata: projects loaded from API: %d\n", len(data["metadata"].(map[string]interface{})["projects"].([]interface{})))
	} else {
		log.Printf("metadata: no projects specified, loading from API\n")
		// 获取组织下的所有项目信息
		projects, err := c.sentryAPI.Projects(c.sentryOrgSlug)
		if err != nil {
			log.Printf("Failed to fetch projects: %v\n", err)
			return nil
		}
		// 更新data中的projects和projects_slug字段
		data["metadata"].(map[string]interface{})["projects"] = projects
		for _, project := range projects {
			projectsSlug := data["metadata"].(map[string]interface{})["projects_slug"].([]string)
			data["metadata"].(map[string]interface{})["projects_slug"] = append(projectsSlug, project.Slug)
			// 获取项目环境信息
			envs, err := c.sentryAPI.Environments(org.Slug, project)
			if err != nil {
				log.Printf("Failed to fetch environments for project %s: %v\n", project.Slug, err)
				continue
			}
			// 更新data中的projects_envs字段
			projectsEnvs := data["metadata"].(map[string]interface{})["projects_envs"].(map[string]interface{})
			projectsEnvs[project.Slug] = envs

			// 构建项目问题数据
			if c.issueMetrics {
				projectsIssueData := make(map[string]map[string]map[string][]interface{})
				for _, env := range envs {
					projectsIssueData[project.Slug] = make(map[string]map[string][]interface{})

					if c.get1hMetrics {
						log.Printf("metadata: getting issues from API - project: %s env: %s age: 1h\n", project.Slug, env)
						issues1h, err := c.sentryAPI.Issues(org.Slug, project, env, "1h")
						if err != nil {
							log.Printf("Failed to fetch issues for project %s, env %s, age 1h: %v\n", project.Slug, env, err)
							continue
						}
						// projectsIssueData[project.Slug][env]["1h"] = issues1h["all"].([]interface{})
						// 确保map的每一层都已初始化
						if _, ok := projectsIssueData[project.Slug]; !ok {
							projectsIssueData[project.Slug] = make(map[string]map[string][]interface{})
						}
						if _, ok := projectsIssueData[project.Slug][env]; !ok {
							projectsIssueData[project.Slug][env] = make(map[string][]interface{})
						}
						// 确保allIssues是一个切片，然后赋值
						allIssues, ok := issues1h["all"].([]interface{})
						if !ok {
							log.Printf("Unexpected type for 'all' key in issues data, expected []interface{} but got %T", issues1h["all"])
							continue
						}
						projectsIssueData[project.Slug][env]["1h"] = allIssues
					}

					if c.get24hMetrics {
						log.Printf("metadata: getting issues from API - project: %s env: %s age: 24h\n", project.Slug, env)
						issues24h, err := c.sentryAPI.Issues(org.Slug, project, env, "24h")
						if err != nil {
							log.Printf("Failed to fetch issues for project %s, env %s, age 24h: %v\n", project.Slug, env, err)
							continue
						}
						// 确保map的每一层都已初始化
						if _, ok := projectsIssueData[project.Slug]; !ok {
							projectsIssueData[project.Slug] = make(map[string]map[string][]interface{})
						}
						if _, ok := projectsIssueData[project.Slug][env]; !ok {
							projectsIssueData[project.Slug][env] = make(map[string][]interface{})
						}
						// 确保allIssues是一个切片，然后赋值
						allIssues, ok := issues24h["all"].([]interface{})
						if !ok {
							log.Printf("Unexpected type for 'all' key in issues data, expected []interface{} but got %T", issues24h["all"])
							continue
						}
						projectsIssueData[project.Slug][env]["24h"] = allIssues
						// projectsIssueData[project.Slug][env]["24h"] = issues24h["all"].([]interface{})
					}

					if c.get14dMetrics {
						log.Printf("metadata: getting issues from API - project: %s env: %s age: 14d\n", project.Slug, env)
						issues14d, err := c.sentryAPI.Issues(org.Slug, project, env, "14d")
						if err != nil {
							log.Printf("Failed to fetch issues for project %s, env %s, age 14d: %v\n", project.Slug, env, err)
							continue
						}
						// projectsIssueData[project.Slug][env]["14d"] = issues14d["all"].([]interface{})
						// 确保map的每一层都已初始化
						if _, ok := projectsIssueData[project.Slug]; !ok {
							projectsIssueData[project.Slug] = make(map[string]map[string][]interface{})
						}
						if _, ok := projectsIssueData[project.Slug][env]; !ok {
							projectsIssueData[project.Slug][env] = make(map[string][]interface{})
						}
						// 确保allIssues是一个切片，然后赋值
						allIssues, ok := issues14d["all"].([]interface{})
						if !ok {
							log.Printf("Unexpected type for 'all' key in issues data, expected []interface{} but got %T", issues14d["all"])
							continue
						}
						projectsIssueData[project.Slug][env]["14d"] = allIssues
					}
				}
				data["projects_data"] = projectsIssueData
			}
		}
		log.Printf("metadata: projects loaded from API: %d\n", len(data["metadata"].(map[string]interface{})["projects"].([]interface{})))
	}
	// 写入缓存
	writeCache(JSONCacheFile, data, time.Now().Add(DefaultCacheExpireTimestamp).Unix())
	return data
}

// buildSentryData 从缓存中读取数据
func (c *SentryCollector) buildSentryData() map[string]interface{} {
	data, err := getCached(JSONCacheFile)
	if err != nil {
		//if c.sentryAPI.liveness() {
		//	log.Printf("cache: %s not found, but API is live. Rebuilding from API...\n", JSONCacheFile)
		//	apiData := c.buildSentryDataFromAPI()
		//	return apiData
		log.Printf("cache: %s not found, but API is not live. Using cached data...\n", JSONCacheFile)
		return nil
	}
	if data == nil {
		log.Printf("cache: %s not found.\n", JSONCacheFile)
		log.Printf("cache: rebuilding from API...\n")
		apiData := c.buildSentryDataFromAPI()
		return apiData
	}
	log.Printf("cache: reading data structure from file: %s\n", JSONCacheFile)
	return data
}

// Describe 方法用于描述所有收集器的指标
func (c *SentryCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect 方法用于收集指标
func (c *SentryCollector) Collect(ch chan<- prometheus.Metric) {
	// 拿到缓存的数据
	data := c.buildSentryData()
	metadata := data["metadata"].(map[string]interface{})
	projectsData, ok := data["projects_data"].(map[string]interface{})
	if !ok {
		log.Println("Type assertion failed for data[\"projects_data\"].")
		return
	}
	// 获取组织和项目信息
	c.org = metadata["org"].(map[string]interface{})
	c.projectsData = projectsData

	// 收集问题指标
	if c.issueMetrics {
		// 创建一个直方图指标，用于记录每个项目及环境的未解决问题数量分布
		issuesHistogramMetrics := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "sentry_open_issues_histogram",
				Help:    "Histogram of open issues (aka is:unresolved) count per project and environment",
				Buckets: []float64{1, 5, 10, 50, 100, 500}, // 自定义的桶边界，根据实际情况调整
			},
			[]string{"project_slug", "environment"},
		)

		log.Printf("collector: loading projects issues\n")
		for _, project := range metadata["projects"].([]interface{}) {
			projectMap := project.(map[string]interface{})
			projectSlug := projectMap["slug"].(string)
			envsInterface := metadata["projects_envs"].(map[string]interface{})[projectSlug].([]interface{})

			// 将 []interface{} 类型的 envsInterface 转换为 []string 类型的 envs
			var envs []string
			for _, env := range envsInterface {
				envs = append(envs, env.(string))
			}

			projectIssues := projectsData[projectSlug].(map[string]interface{})
			for _, env := range envs {
				log.Printf("collector: loading issues - project: %s env: %s\n", projectSlug, env)
				projectIssuesEnvInterface := projectIssues[env]
				if projectIssuesEnvInterface == nil {
					log.Printf("No issues data for project: %s env: %s\n", projectSlug, env)
					continue
				}

				projectIssuesEnv := projectIssuesEnvInterface.(map[string]interface{})
				//fmt.Println(projectIssuesEnv)

				projectIssues1hInterface := projectIssuesEnv["1h"]
				projectIssues24hInterface := projectIssuesEnv["24h"]
				projectIssues14dInterface := projectIssuesEnv["14d"]

				if projectIssues1hInterface != nil {
					projectIssues1h := projectIssues1hInterface.([]interface{})
					events1h := 0
					for _, issue := range projectIssues1h {
						issueMap := issue.(map[string]interface{})
						switch v := issueMap["count"].(type) {
						case float64:
							events1h += int(v)
						case string:
							floatVal, err := strconv.ParseFloat(v, 64)
							if err != nil {
								log.Printf("Failed to parse count as float64: %v\n", err)
								continue
							}
							events1h += int(floatVal)
						default:
							log.Printf("Unexpected type for 'count': %T\n", v)
							continue
						}
					}
					issuesHistogramMetrics.WithLabelValues(
						projectSlug,
						env,
					).Observe(float64(events1h))
				} else {
					log.Printf("No 1h issues data for project: %s env: %s\n", projectSlug, env)
				}

				if projectIssues24hInterface != nil {
					projectIssues24h := projectIssues24hInterface.([]interface{})
					events24h := 0
					for _, issue := range projectIssues24h {
						issueMap := issue.(map[string]interface{})
						switch v := issueMap["count"].(type) {
						case float64:
							events24h += int(v)
						case string:
							floatVal, err := strconv.ParseFloat(v, 64)
							if err != nil {
								log.Printf("Failed to parse count as float64: %v\n", err)
								continue
							}
							events24h += int(floatVal)
						default:
							log.Printf("Unexpected type for 'count': %T\n", v)
							continue
						}
					}
					issuesHistogramMetrics.WithLabelValues(
						projectSlug,
						env,
					).Observe(float64(events24h))

				} else {
					log.Printf("No 24h issues data for project: %s env: %s\n", projectSlug, env)
				}

				if projectIssues14dInterface != nil {
					projectIssues14d := projectIssues14dInterface.([]interface{})
					events14d := 0
					for _, issue := range projectIssues14d {
						issueMap := issue.(map[string]interface{})
						switch v := issueMap["count"].(type) {
						case float64:
							events14d += int(v)
						case string:
							floatVal, err := strconv.ParseFloat(v, 64)
							if err != nil {
								log.Printf("Failed to parse count as float64: %v\n", err)
								continue
							}
							events14d += int(floatVal)
						default:
							log.Printf("Unexpected type for 'count': %T\n", v)
							continue
						}
					}

					issuesHistogramMetrics.WithLabelValues(
						projectSlug,
						env,
					).Observe(float64(events14d))
				} else {
					log.Printf("No 14d issues data for project: %s env: %s\n", projectSlug, env)
				}
			}
		}

		issuesHistogramMetrics.Collect(ch)
	}

	// 收集 open issue events 指标
	if c.issueMetrics {
		issuesMetrics := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sentry_open_issue_events",
				Help: "Number of open issues (aka is:unresolved) per project",
			},
			[]string{
				"issue_id",
				"logger",
				"level",
				"status",
				"platform",
				"project_slug",
				"environment",
				"release",
				"isUnhandled",
				"firstSeen",
				"lastSeen",
			},
		)

		for _, project := range metadata["projects"].([]interface{}) {
			projectMap := project.(map[string]interface{})
			projectSlug := projectMap["slug"].(string)

			envsInterface, ok := metadata["projects_envs"].(map[string]interface{})[projectSlug].([]interface{})
			if !ok {
				log.Printf("Failed to assert environments as []interface{}: %v\n", metadata["projects_envs"].(map[string]interface{})[projectSlug])
				continue
			}

			var envs []string
			for _, env := range envsInterface {
				envStr, ok := env.(string)
				if !ok {
					log.Printf("Failed to assert environment as string: %v\n", env)
					continue
				}
				envs = append(envs, envStr)
			}

			projectIssues := projectsData[projectSlug].(map[string]interface{})
			for _, env := range envs {
				projectIssuesEnvInterface := projectIssues[env]
				if projectIssuesEnvInterface == nil {
					log.Printf("No issues data for project: %s env: %s\n", projectSlug, env)
					continue
				}
				projectIssuesEnv := projectIssuesEnvInterface.(map[string]interface{})

				timeFrames := []string{"1h", "24h", "14d"}
				for _, timeFrame := range timeFrames {
					projectIssuesTimeFrameInterface, ok := projectIssuesEnv[timeFrame]
					if !ok {
						log.Printf("No %s issues data for project: %s env: %s\n", timeFrame, projectSlug, env)
						continue
					}
					projectIssuesTimeFrame := projectIssuesTimeFrameInterface.([]interface{})
					for _, issue := range projectIssuesTimeFrame {
						issueMap := issue.(map[string]interface{})
						release, err := c.sentryAPI.IssueRelease(issueMap["id"].(string), env)
						if err != nil {
							log.Printf("Failed to fetch release for issue %s: %v\n", issueMap["id"].(string), err)
							continue
						}

						count, ok := issueMap["count"].(float64)
						if !ok {
							countStr, ok := issueMap["count"].(string)
							if ok {
								count, err = strconv.ParseFloat(countStr, 64)
								if err != nil {
									log.Printf("Failed to parse count as float64: %v\n", err)
									continue
								}
							} else {
								log.Printf("Unexpected type for 'count': %T\n", issueMap["count"])
								continue
							}
						}

						id, ok := issueMap["id"].(string)
						if !ok {
							log.Println("Type assertion failed for issueMap[\"id\"].")
							return
						}
						logger, ok := issueMap["logger"].(string)
						if !ok {
							log.Println("Type assertion failed for issueMap[\"logger\"].")
							return
						}
						level, ok := issueMap["level"].(string)
						if !ok {
							log.Println("Type assertion failed for issueMap[\"level\"].")
							return
						}
						status, ok := issueMap["status"].(string)
						if !ok {
							log.Println("Type assertion failed for issueMap[\"status\"].")
							return
						}
						platform, ok := issueMap["platform"].(string)
						if !ok {
							log.Println("Type assertion failed for issueMap[\"platform\"].")
							return
						}
						isUnhandled, ok := issueMap["isUnhandled"]
						if !ok {
							log.Println("Key 'isUnhandled' not found in issueMap.")
							return
						}
						firstSeen, ok := issueMap["firstSeen"]
						if !ok {
							log.Println("Key 'firstSeen' not found in issueMap.")
							return
						}
						lastSeen, ok := issueMap["lastSeen"]
						if !ok {
							log.Println("Key 'lastSeen' not found in issueMap.")
							return
						}

						issuesMetrics.WithLabelValues(
							id,
							logger,
							level,
							status,
							platform,
							projectSlug,
							env,
							release,
							fmt.Sprintf("%v", isUnhandled),
							fmt.Sprintf("%v", firstSeen),
							fmt.Sprintf("%v", lastSeen),
						).Set(count)
					}
				}
			}
		}

		issuesMetrics.Collect(ch)
	}

	// 收集 events 指标
	if c.eventsMetrics {
		projectEventsMetrics := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sentry_events",
				Help: "Total events counts per project",
			},
			[]string{"project_slug", "stat"},
		)

		for _, project := range metadata["projects"].([]interface{}) {
			projectMap := project.(map[string]interface{})
			projectSlug := projectMap["slug"].(string)
			events, err := c.sentryAPI.ProjectStats(c.sentryOrgSlug, projectSlug)
			if err != nil {
				log.Printf("Failed to fetch project stats for project %s: %v\n", projectSlug, err)
				continue
			}
			for stat, value := range events {
				projectEventsMetrics.WithLabelValues(
					projectSlug,
					stat,
				).Add(float64(value))
			}
		}

		projectEventsMetrics.Collect(ch)
	}

	// 收集 rate limit 指标
	if c.rateLimitMetrics {
		projectRateMetrics := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sentry_rate_limit_events_sec",
				Help: "Rate limit events per second for a project",
			},
			[]string{"project_slug"},
		)

		for _, project := range metadata["projects"].([]interface{}) {
			projectMap := project.(map[string]interface{})
			projectSlug := projectMap["slug"].(string)
			rateLimitSecond, err := c.sentryAPI.RateLimit(c.sentryOrgSlug, projectSlug)
			if err != nil {
				log.Printf("Failed to fetch rate limit for project %s: %v\n", projectSlug, err)
				continue
			}
			projectRateMetrics.WithLabelValues(
				projectSlug,
			).Set(rateLimitSecond)
		}

		projectRateMetrics.Collect(ch)
	}
}
