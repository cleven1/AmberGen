package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// NewsSearcher 新闻搜索工具
type NewsSearcher struct {
	*BaseTool
	apiKey string
	client *http.Client
}

type NewsResponse struct {
	Code  int    `json:"code,omitempty"`
	LogID string `json:"log_id,omitempty"`
	Msg   any    `json:"msg,omitempty"`
	Data  struct {
		Type         string `json:"_type,omitempty"`
		QueryContext struct {
			OriginalQuery string `json:"originalQuery,omitempty"`
		} `json:"queryContext,omitempty"`
		WebPages struct {
			WebSearchURL          string `json:"webSearchUrl,omitempty"`
			TotalEstimatedMatches int    `json:"totalEstimatedMatches,omitempty"`
			Value                 []struct {
				ID               any       `json:"id,omitempty"`
				Name             string    `json:"name,omitempty"`
				URL              string    `json:"url,omitempty"`
				DisplayURL       string    `json:"displayUrl,omitempty"`
				Snippet          string    `json:"snippet,omitempty"`
				DateLastCrawled  time.Time `json:"dateLastCrawled,omitempty"`
				CachedPageURL    any       `json:"cachedPageUrl,omitempty"`
				Language         any       `json:"language,omitempty"`
				IsFamilyFriendly any       `json:"isFamilyFriendly,omitempty"`
				IsNavigational   any       `json:"isNavigational,omitempty"`
			} `json:"value,omitempty"`
			SomeResultsRemoved bool `json:"someResultsRemoved,omitempty"`
		} `json:"webPages,omitempty"`
		Images struct {
			ID           any `json:"id,omitempty"`
			ReadLink     any `json:"readLink,omitempty"`
			WebSearchURL any `json:"webSearchUrl,omitempty"`
			Value        []struct {
				WebSearchURL       any    `json:"webSearchUrl,omitempty"`
				Name               any    `json:"name,omitempty"`
				ThumbnailURL       string `json:"thumbnailUrl,omitempty"`
				DatePublished      any    `json:"datePublished,omitempty"`
				ContentURL         string `json:"contentUrl,omitempty"`
				HostPageURL        string `json:"hostPageUrl,omitempty"`
				ContentSize        any    `json:"contentSize,omitempty"`
				EncodingFormat     any    `json:"encodingFormat,omitempty"`
				HostPageDisplayURL any    `json:"hostPageDisplayUrl,omitempty"`
				Width              int    `json:"width,omitempty"`
				Height             int    `json:"height,omitempty"`
				Thumbnail          any    `json:"thumbnail,omitempty"`
			} `json:"value,omitempty"`
			IsFamilyFriendly any `json:"isFamilyFriendly,omitempty"`
		} `json:"images,omitempty"`
		Videos struct {
			ID               any   `json:"id,omitempty"`
			ReadLink         any   `json:"readLink,omitempty"`
			WebSearchURL     any   `json:"webSearchUrl,omitempty"`
			IsFamilyFriendly any   `json:"isFamilyFriendly,omitempty"`
			Scenario         any   `json:"scenario,omitempty"`
			Value            []any `json:"value,omitempty"`
		} `json:"videos,omitempty"`
	} `json:"data,omitempty"`
}

func NewNewsSearcher(apiKey string) *NewsSearcher {
	tool := &NewsSearcher{
		BaseTool: NewBaseTool("news_searcher", "搜索最新新闻并提供摘要"),
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 10 * time.Second},
	}

	tool.AddParameter("query", ParameterSpec{
		Type:        "string",
		Description: "搜索关键词",
		Required:    true,
	})
	return tool
}

func (n *NewsSearcher) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("invalid or missing query parameter")
	}

	// 构建API URL
	apiURL := "https://api.bochaai.com/v1/web-search"
	data := map[string]interface{}{
		"query": query,
	}

	// 将数据编码为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("JSON marshal failed: %v\n", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Create request failed: %v\n", err)
	}

	req.Header.Set("Authorization", "Bearer "+n.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := n.client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var newsResp NewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		fmt.Printf("Decode response failed: %v\n", err)
	}

	// 生成摘要
	summary := n.generateSummary(newsResp)
	return summary, nil
}

func (n *NewsSearcher) generateSummary(articles NewsResponse) string {
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("找到 %d 条相关新闻：\n\n", len(articles.Data.WebPages.Value)))

	for i, article := range articles.Data.WebPages.Value {
		if i >= 5 { // 只显示前5条
			break
		}
		summary.WriteString(fmt.Sprintf("%d. %s\n", i+1, article.Name))
		summary.WriteString(fmt.Sprintf("   发布时间: %s\n", article.DateLastCrawled.Format("2006-01-02 15:04:05")))
		summary.WriteString(fmt.Sprintf("   摘要: %s\n\n", article.Snippet))
	}

	return summary.String()
}
