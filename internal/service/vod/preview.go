package vod

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type M3U8Fetcher interface {
	Get(ctx context.Context, url string) (string, error)
}

type httpM3U8Fetcher struct{}

func (httpM3U8Fetcher) Get(ctx context.Context, rawURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://example.com/")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch m3u8 status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *ListingService) Preview(ctx context.Context, vodID int) (string, error) {
	row, err := s.store.VODByID(ctx, vodID)
	if err != nil {
		return "", fmt.Errorf("get vod: %w", err)
	}
	if len(row) == 0 || atoi(str(row["showtype"])) > 0 || strings.TrimSpace(str(row["play_url"])) == "" {
		return "", nil
	}
	servers, err := s.store.Servers(ctx)
	if err != nil {
		return "", fmt.Errorf("list servers: %w", err)
	}
	m3u8URL := previewSourceURL(row, servers)
	if m3u8URL == "" {
		return "", nil
	}
	body := s.generatePreviewM3U8(ctx, m3u8URL, 0, 300)
	if body == "" {
		return "#EXT-X-ENDLIST", nil
	}
	return body, nil
}

func previewSourceURL(row map[string]interface{}, servers []map[string]interface{}) string {
	httpURL := cleanCover(str(row["play_url"]))
	if httpURL == "" {
		return ""
	}
	if strings.HasPrefix(httpURL, "http://") || strings.HasPrefix(httpURL, "https://") {
		return httpURL
	}
	host := ""
	playServerID := atoi(str(row["play_srvid"]))
	for _, server := range servers {
		if str(server["srvtype"]) == "play" && atoi(str(server["srvid"])) == playServerID {
			host = strings.TrimRight(str(server["srvhost"]), "/")
			break
		}
	}
	if host == "" {
		return ""
	}
	return host + "/" + strings.TrimLeft(httpURL, "/")
}

func (s *ListingService) generatePreviewM3U8(ctx context.Context, m3u8URL string, startTime float64, endTime float64) string {
	content, err := s.fetcher.Get(ctx, m3u8URL)
	if err != nil || content == "" {
		return ""
	}
	subURL := subM3U8URL(m3u8URL, content)
	if subURL == "" {
		return ""
	}
	subContent, err := s.fetcher.Get(ctx, subURL)
	if err != nil || subContent == "" {
		return ""
	}
	return processPreviewM3U8(subURL, subContent, startTime, endTime)
}

func subM3U8URL(baseURL string, content string) string {
	domain := urlDomain(baseURL)
	if domain == "" {
		return ""
	}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
				return line
			}
			return strings.TrimRight(domain, "/") + "/" + strings.TrimLeft(line, "/")
		}
	}
	return ""
}

func processPreviewM3U8(baseURL string, content string, startTime float64, endTime float64) string {
	domain := urlDomain(baseURL)
	lines := strings.Split(content, "\n")
	out := []string{}
	currentDuration := 0.0
	startFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "#EXT-X-KEY") {
			line = rewriteKeyURI(line, domain)
		}
		if strings.HasPrefix(line, "#EXTM3U") || strings.HasPrefix(line, "#EXT-X") {
			out = append(out, line)
			continue
		}
		if strings.HasPrefix(line, "#EXTINF:") {
			duration := extinfDuration(line)
			if currentDuration+duration < startTime {
				currentDuration += duration
				continue
			}
			if !startFound {
				startFound = true
			}
			if currentDuration >= endTime {
				break
			}
			out = append(out, line)
			currentDuration += duration
			continue
		}
		if startFound && currentDuration <= endTime {
			if line != "" && !strings.HasPrefix(line, "http") {
				line = strings.TrimRight(domain, "/") + "/" + strings.TrimLeft(line, "/")
			}
			out = append(out, line)
		}
	}
	out = append(out, "#EXT-X-ENDLIST")
	return strings.Join(out, "\n")
}

func rewriteKeyURI(line string, domain string) string {
	const marker = `URI="`
	start := strings.Index(line, marker)
	if start < 0 {
		return line
	}
	start += len(marker)
	end := strings.Index(line[start:], `"`)
	if end < 0 {
		return line
	}
	end += start
	raw := line[start:end]
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return line
	}
	return line[:start] + strings.TrimRight(domain, "/") + "/" + strings.TrimLeft(raw, "/") + line[end:]
}

func extinfDuration(line string) float64 {
	value := strings.TrimPrefix(line, "#EXTINF:")
	value = strings.TrimSuffix(value, ",")
	var duration float64
	_, _ = fmt.Sscan(value, &duration)
	return duration
}

func urlDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}
