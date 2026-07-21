package aiundress

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type R2UploaderConfig struct {
	AccountID string
	AccessKey string
	SecretKey string
	Bucket    string
	Timeout   time.Duration
}

func (c R2UploaderConfig) Configured() bool {
	return strings.TrimSpace(c.AccountID) != "" &&
		strings.TrimSpace(c.AccessKey) != "" &&
		strings.TrimSpace(c.SecretKey) != "" &&
		strings.TrimSpace(c.Bucket) != ""
}

type R2Uploader struct {
	cfg    R2UploaderConfig
	client *http.Client
	now    func() time.Time
}

func NewR2Uploader(cfg R2UploaderConfig) *R2Uploader {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &R2Uploader{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
		now:    time.Now,
	}
}

func (u *R2Uploader) Upload(ctx context.Context, localPath string, objectKey string) error {
	if u == nil || !u.cfg.Configured() {
		return fmt.Errorf("r2 config incomplete")
	}
	body, err := os.ReadFile(localPath)
	if err != nil {
		return err
	}

	accountID := strings.TrimSpace(u.cfg.AccountID)
	accessKey := strings.TrimSpace(u.cfg.AccessKey)
	secretKey := strings.TrimSpace(u.cfg.SecretKey)
	bucket := strings.Trim(strings.TrimSpace(u.cfg.Bucket), "/")
	key := strings.TrimLeft(objectKey, "/")
	host := accountID + ".r2.cloudflarestorage.com"
	canonicalURI := "/" + bucket + "/" + encodeObjectKey(key)
	endpointURL := "https://" + host + canonicalURI

	payloadHash := sha256Hex(body)
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(localPath)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	now := u.now().UTC()
	amzDate := now.Format("20060102T150405Z")
	shortDate := now.Format("20060102")
	region := "auto"
	service := "s3"
	scope := shortDate + "/" + region + "/" + service + "/aws4_request"

	canonicalHeaders := "content-type:" + contentType + "\n" +
		"host:" + host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"
	canonicalRequest := strings.Join([]string{
		http.MethodPut,
		canonicalURI,
		"",
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signingKey := awsSigningKey(secretKey, shortDate, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	authorization := "AWS4-HMAC-SHA256 Credential=" + accessKey + "/" + scope +
		", SignedHeaders=" + signedHeaders + ", Signature=" + signature

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpointURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Host = host
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	req.Header.Set("Authorization", authorization)
	req.ContentLength = int64(len(body))

	resp, err := u.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("r2 upload status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
}

func encodeObjectKey(key string) string {
	parts := strings.Split(key, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func awsSigningKey(secret string, date string, region string, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(value))
	return mac.Sum(nil)
}
