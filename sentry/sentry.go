package sentry

import (
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type SentryAPI struct {
	BaseURL   string
	AuthToken string
	Client    *http.Client
}

type Organization struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Status Status `json:"status"`
}

// Status 结构体定义
type Status struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Platform string `json:"platform"`
}

//type Event struct {
//	Timestamp string `json:"timestamp"`
//	Count     int    `json:"count"`
//}

type Event []interface{}

func NewSentryAPI(baseURL, authToken string) *SentryAPI {
	return &SentryAPI{
		BaseURL:   baseURL,
		AuthToken: authToken,
		Client:    &http.Client{},
	}
}

// Get 发送请求验证Token
func (s *SentryAPI) Get(url string) (*http.Response, error) {

	// fmt.Println(s.BaseURL + url)
	req, err := http.NewRequest(http.MethodGet, s.BaseURL+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.AuthToken)

	// 使用闭包传递局部变量 resp 和 err
	var resp *http.Response
	var doErr error
	operation := func() error {
		resp, doErr = s.Client.Do(req)
		if doErr != nil {
			return doErr
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP error: %s", resp.Status)
		}
		return nil
	}

	// 使用 retry-go 实现重试逻辑
	err = retry.Do(
		operation,
		retry.Attempts(3),              // 设置重试次数
		retry.Delay(2*time.Second),     // 设置重试间隔
		retry.MaxDelay(10*time.Second), // 设置最大重试间隔
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Attempt %d: %v\n", n, err)
		}),
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Organizations 获取组织列表
func (s *SentryAPI) Organizations() ([]Organization, error) {
	resp, err := s.Get("organizations/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var orgs []Organization
	err = json.NewDecoder(resp.Body).Decode(&orgs)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetOrg 获取单个组织
func (s *SentryAPI) GetOrg(orgSlug string) (*Organization, error) {
	resp, err := s.Get("organizations/" + orgSlug + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %v", err)
	}
	defer resp.Body.Close()
	// 打印接口返回的响应内容
	//fmt.Println("Response Status:", resp.Status)
	//fmt.Println("Response Headers:", resp.Header)
	// 读取响应Body内容
	body, err := ioutil.ReadAll(resp.Body)
	// 解码 JSON 到 Organization 结构体
	var org Organization
	err = json.Unmarshal(body, &org)
	if err != nil {
		return nil, fmt.Errorf("failed to decode organization JSON: %v", err)
	}
	// 打印解码后的组织信息
	fmt.Println("Organization ID:", org.ID)
	fmt.Println("Organization Slug:", org.Slug)
	fmt.Println("Organization Name:", org.Name)
	fmt.Println("Organization Status ID:", org.Status.ID)
	fmt.Println("Organization Status Name:", org.Status.Name)

	return &org, nil
}

// Projects 获取组织下的项目列表
func (s *SentryAPI) Projects(orgSlug string) ([]Project, error) {
	resp, err := s.Get(fmt.Sprintf("organizations/%s/projects/?all_projects=1", orgSlug))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var projects []Project
	err = json.NewDecoder(resp.Body).Decode(&projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject 获取单个项目
func (s *SentryAPI) GetProject(orgSlug, projectSlug string) (*Project, error) {
	resp, err := s.Get(fmt.Sprintf("projects/%s/%s/", orgSlug, projectSlug))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var project Project
	//err = json.NewDecoder(resp.Body).Decode(&project)
	err = json.Unmarshal(body, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to decode project JSON: %v", err)
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ProjectStats 获取项目统计信息
func (s *SentryAPI) ProjectStats(orgSlug, projectSlug string) (map[string]int, error) {
	firstDayMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1).Unix()
	today := time.Now().Unix()
	statNames := []string{"received", "rejected", "blacklisted"}
	stats := make(map[string][]Event)
	projectEvents := make(map[string]int)

	for _, statName := range statNames {
		resp, err := s.Get(fmt.Sprintf("projects/%s/%s/stats/?stat=%s&since=%d&until=%d", orgSlug, projectSlug, statName, firstDayMonth, today))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		var events []Event
		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &events)
		if err != nil {
			return nil, err
		}
		stats[statName] = events
	}
	for statName, events := range stats {
		eventsCount := 0
		for _, event := range events {
			eventsCount += int(event[1].(float64))
		}
		projectEvents[statName] = eventsCount
	}

	return projectEvents, nil
}

// Environments 获取项目环境列表
func (s *SentryAPI) Environments(orgSlug string, project Project) ([]string, error) {
	resp, err := s.Get(fmt.Sprintf("projects/%s/%s/environments/", orgSlug, project.Slug))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var environments []struct {
		Name string `json:"name"`
	}
	err = json.NewDecoder(resp.Body).Decode(&environments)
	if err != nil {
		return nil, err
	}
	envs := make([]string, len(environments))
	for i, env := range environments {
		envs[i] = env.Name
	}
	return envs, nil
}

// Issues 获取项目问题列表
func (s *SentryAPI) Issues(orgSlug string, project Project, environment string, age string) (map[string]interface{}, error) {
	issuesURL := fmt.Sprintf("projects/%s/%s/issues/?project=%s&sort=date&query=age%%3A-%s", orgSlug, project.Slug, project.ID, age)

	if environment != "" {
		issuesURL = fmt.Sprintf("%s&environment=%s", issuesURL, environment)
	}

	resp, err := s.Get(issuesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issues []interface{}
	err = json.NewDecoder(resp.Body).Decode(&issues)
	if err != nil {
		return nil, err
	}

	// 明确处理空issues情况
	if len(issues) == 0 {
		return map[string]interface{}{"message": "No issues found", "all": []interface{}{}}, nil
	}

	result := map[string]interface{}{"all": issues}
	if environment != "" {
		result[environment] = issues
	}
	return result, nil
}

// Events 获取项目事件列表
func (s *SentryAPI) Events(orgSlug string, project Project, environment string) (map[string]interface{}, error) {
	eventsURL := fmt.Sprintf("projects/%s/%s/events/?project=%s&sort=date", orgSlug, project.Slug, project.ID)

	if environment != "" {
		eventsURL = fmt.Sprintf("%s&environment=%s", eventsURL, environment)
	}
	resp, err := s.Get(eventsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var events []interface{}
	err = json.NewDecoder(resp.Body).Decode(&events)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{"all": events}
	if environment != "" {
		result[environment] = events
	}
	return result, nil
}

// IssueEvents 获取问题事件列表
func (s *SentryAPI) IssueEvents(issueID string, environment string) (map[string]interface{}, error) {
	issueEventsURL := fmt.Sprintf("issues/%s/events/", issueID)

	if environment != "" {
		issueEventsURL = fmt.Sprintf("%s&environment=%s&sort=date", issueEventsURL, environment)
	}
	resp, err := s.Get(issueEventsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issueEvents []interface{}
	err = json.NewDecoder(resp.Body).Decode(&issueEvents)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{"all": issueEvents}
	if environment != "" {
		result[environment] = issueEvents
	}
	return result, nil
}

// IssueRelease 获取问题发布版本
func (s *SentryAPI) IssueRelease(issueID string, environment string) (string, error) {
	issueReleaseURL := fmt.Sprintf("issues/%s/current-release/", issueID)

	if environment != "" {
		issueReleaseURL = fmt.Sprintf("%s?environment=%s", issueReleaseURL, environment)
	}

	resp, err := s.Get(issueReleaseURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		CurrentRelease struct {
			Release struct {
				Version string `json:"version"`
			} `json:"release"`
		} `json:"currentRelease"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.CurrentRelease.Release.Version, nil
}

func (s *SentryAPI) ProjectReleases(orgSlug string, project Project, environment string) (map[string]interface{}, error) {
	projReleasesURL := fmt.Sprintf("organizations/%s/releases/?project=%s&sort=date", orgSlug, project.ID)

	if environment != "" {
		projReleasesURL = fmt.Sprintf("%s&environment=%s", projReleasesURL, environment)
	}

	resp, err := s.Get(projReleasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project releases: %v", err)
	}
	defer resp.Body.Close()

	var releases []interface{}
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return nil, fmt.Errorf("failed to decode project releases JSON: %v", err)
	}

	releasesMap := make(map[string]interface{})
	if environment != "" {
		releasesMap[environment] = releases
	} else {
		releasesMap["all"] = releases
	}

	return releasesMap, nil
}

// rateLimit 获取项目速率限制
func (s *SentryAPI) RateLimit(orgSlug, projectSlug string) (float64, error) {
	rateLimitURL := fmt.Sprintf("projects/%s/%s/keys/", orgSlug, projectSlug)

	resp, err := s.Get(rateLimitURL)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch rate limit: %v", err)
	}
	defer resp.Body.Close()

	var rateLimits []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&rateLimits)
	if err != nil {
		return 0, fmt.Errorf("failed to decode rate limit JSON: %v", err)
	}

	if len(rateLimits) > 0 && rateLimits[0]["rateLimit"] != nil {
		rateLimit := rateLimits[0]["rateLimit"].(map[string]interface{})
		window := rateLimit["window"].(float64)
		count := rateLimit["count"].(float64)
		if window != 0 {
			rateLimitSecond := count / window
			return rateLimitSecond, nil
		}
	}

	return 0, nil
}

///////

func (s *SentryAPI) liveness() bool {
	// 占位函数，用于检查应用程序是否正常运行
	return true // TODO: 实现实际的健康检查逻辑
}

func (s *SentryAPI) readiness() error {
	// 检查 SentryAPI 实例是否准备好接收请求
	api := NewSentryAPI(s.BaseURL, s.AuthToken)
	_, err := api.ProjectReleases("example_org_slug", Project{ID: "example_project_id", Slug: "example_project_slug"}, "")
	if err != nil {
		return fmt.Errorf("SentryAPI 就绪检查失败: %v", err)
	}
	return nil
}
