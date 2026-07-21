package aiundress

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"xj_comp/internal/domain"
	userRepo "xj_comp/internal/repository/user"
	vodService "xj_comp/internal/service/vod"
)

type AuthStore interface {
	UserBySession(ctx context.Context, sid string) (map[string]interface{}, error)
}

type Store interface {
	Count(ctx context.Context, uid int, module int) (int, error)
	List(ctx context.Context, uid int, module int, total int, page int, pageSize int) ([]map[string]interface{}, error)
	ByID(ctx context.Context, id int) (map[string]interface{}, error)
	ByUIDImage(ctx context.Context, uid int, image string) (map[string]interface{}, error)
	CreateUpload(ctx context.Context, uid int, image string, now int64) (int, error)
	RefreshUpload(ctx context.Context, id int, now int64) error
	MarkDeleted(ctx context.Context, id int, updateTime int64) error
	SettingByUUID(ctx context.Context, uuid string) (string, error)
}

type ExternalClient interface {
	PostJSON(ctx context.Context, path string, payload map[string]interface{}) (ExternalResponse, error)
}

type FileDeleter interface {
	Delete(path string) error
}

type ObjectUploader interface {
	Upload(ctx context.Context, localPath string, objectKey string) error
}

type osFileDeleter struct{}

func (osFileDeleter) Delete(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

type noopObjectUploader struct{}

func (noopObjectUploader) Upload(context.Context, string, string) error {
	return nil
}

type ExternalResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ResourceListInput = domain.AIUndressResourceListInput

type HTTPExternalClient struct {
	baseURL string
	key     string
	client  *http.Client
}

func NewHTTPExternalClient(host string, key string, timeout time.Duration) *HTTPExternalClient {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPExternalClient{
		baseURL: "https://" + strings.TrimRight(host, "/"),
		key:     strings.TrimSpace(key),
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *HTTPExternalClient) PostJSON(ctx context.Context, path string, payload map[string]interface{}) (ExternalResponse, error) {
	if c == nil || c.client == nil || c.baseURL == "https://" || c.key == "" {
		return ExternalResponse{}, fmt.Errorf("aiundress external config incomplete")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ExternalResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return ExternalResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("channel_key", c.key)
	resp, err := c.client.Do(req)
	if err != nil {
		return ExternalResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ExternalResponse{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ExternalResponse{}, err
	}
	var decoded ExternalResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return ExternalResponse{}, err
	}
	return decoded, nil
}

type Service struct {
	auth            AuthStore
	store           Store
	resourceBaseURL string
	uploadPath      string
	env             string
	externalClient  ExternalClient
	fileDeleter     FileDeleter
	uploader        ObjectUploader
	now             func() time.Time
}

func NewService(auth AuthStore, store Store, resourceBaseURL string, env ...string) *Service {
	envValue := ""
	if len(env) > 0 {
		envValue = env[0]
	}
	return &Service{
		auth:            auth,
		store:           store,
		resourceBaseURL: strings.TrimRight(resourceBaseURL, "/"),
		env:             strings.ToLower(strings.TrimSpace(envValue)),
		fileDeleter:     osFileDeleter{},
		uploader:        noopObjectUploader{},
		now:             time.Now,
	}
}

func (s *Service) WithExternalClient(client ExternalClient) *Service {
	s.externalClient = client
	return s
}

func (s *Service) WithUploadPath(uploadPath string) *Service {
	s.uploadPath = strings.TrimRight(strings.TrimSpace(uploadPath), "/")
	return s
}

func (s *Service) WithFileDeleter(deleter FileDeleter) *Service {
	if deleter != nil {
		s.fileDeleter = deleter
	}
	return s
}

func (s *Service) WithObjectUploader(uploader ObjectUploader) *Service {
	if uploader != nil {
		s.uploader = uploader
	}
	return s
}

func (s *Service) Listing(ctx context.Context, token string, page int, module int) (domain.AIUndressListingData, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return domain.AIUndressListingData{}, -1, "获取AI记录失败", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return domain.AIUndressListingData{}, -1, "请先登录", nil
	}
	const pageSize = 10
	total, err := s.store.Count(ctx, uid, module)
	if err != nil {
		return domain.AIUndressListingData{}, -1, "获取AI记录失败", err
	}
	rows, err := s.store.List(ctx, uid, module, total, page, pageSize)
	if err != nil {
		return domain.AIUndressListingData{}, -1, "获取AI记录失败", err
	}
	baseURL, err := s.aiResourceBaseURL(ctx)
	if err != nil {
		return domain.AIUndressListingData{}, -1, "获取AI记录失败", err
	}
	for _, row := range rows {
		row["image"] = s.resourceURL(baseURL, row["image"])
		row["output"] = s.resourceURL(baseURL, row["output"])
	}
	return domain.AIUndressListingData{
		Rows:     rows,
		PageInfo: vodService.PageInfo(total, pageSize, page, "/aiundress/listing?page=[?]"),
	}, 0, "", nil
}

func (s *Service) Upload(ctx context.Context, token string, file multipart.File, header *multipart.FileHeader) (map[string]interface{}, int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return nil, -1, "请先登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -1, "请先登录", nil
	}
	if file == nil || header == nil {
		return nil, -1, "请选择上传图片", nil
	}
	if s.store == nil {
		return nil, -1, "上传失败，请稍后再试", nil
	}

	fileInfo, localPath, image, retcode, errmsg, err := s.saveUploadFile(uid, file, header)
	if err != nil || retcode != 0 {
		return nil, retcode, errmsg, err
	}
	if s.uploader != nil {
		if err := s.uploader.Upload(ctx, localPath, image); err != nil {
			return nil, -1, "上传失败，请稍后再试", err
		}
	}

	now := s.now().Unix()
	row, err := s.store.ByUIDImage(ctx, uid, image)
	if err != nil {
		return nil, -1, "上传失败，请稍后再试", err
	}
	if len(row) == 0 {
		if _, err := s.store.CreateUpload(ctx, uid, image, now); err != nil {
			return nil, -1, "上传失败，请稍后再试", err
		}
	} else if err := s.store.RefreshUpload(ctx, atoi(row["id"]), now); err != nil {
		return nil, -1, "上传失败，请稍后再试", err
	}

	fileInfo["uri"] = image
	return map[string]interface{}{"file": fileInfo}, 0, "", nil
}

func (s *Service) UndressEdge(ctx context.Context, token string, uri string, module int) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "请先登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return -1, "请先登录", nil
	}
	if module <= 0 || module > 6 {
		module = 4
	}
	row, err := s.store.ByUIDImage(ctx, uid, uri)
	if err != nil {
		return -1, "AI 生成失败", err
	}
	if len(row) == 0 {
		return -1, "无效图片路径", nil
	}
	return -1, "AI 生成成功分支暂未迁移", nil
}

func (s *Service) DeleteEdge(ctx context.Context, token string, id int) (int, string, error) {
	user, err := s.userByToken(ctx, token)
	if err != nil {
		return -1, "请先登录", err
	}
	if atoi(user["uid"]) == 0 {
		return -1, "请先登录", nil
	}
	row, err := s.store.ByID(ctx, id)
	if err != nil {
		return -1, "AI 删除失败", err
	}
	if len(row) == 0 {
		return 0, "", nil
	}
	if s.fileDeleter != nil && s.uploadPath != "" {
		if image := strings.TrimSpace(fmt.Sprint(row["image"])); image != "" {
			_ = s.fileDeleter.Delete(filepath.Join(s.uploadPath, image))
		}
		if output := strings.TrimSpace(fmt.Sprint(row["output"])); output != "" {
			_ = s.fileDeleter.Delete(filepath.Join(s.uploadPath, output))
		}
	}
	if err := s.store.MarkDeleted(ctx, atoi(row["id"]), s.now().Unix()); err != nil {
		return -1, "AI 删除失败", err
	}
	return 0, "", nil
}

func (s *Service) saveUploadFile(uid int, file multipart.File, header *multipart.FileHeader) (map[string]interface{}, string, string, int, string, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	if !allowedUploadSuffix(ext) {
		return nil, "", "", -1, fmt.Sprintf("%s:系统只允许上传[jpg, jpeg, gif, png]格式的文件", header.Filename), nil
	}
	if header.Size <= 0 {
		return nil, "", "", -1, "请选择上传图片", nil
	}
	if header.Size > 5*1024*1024 {
		return nil, "", "", -1, fmt.Sprintf("%s:系统限定上传文件不能大于[5120K]", header.Filename), nil
	}
	uploadRoot := strings.TrimSpace(s.uploadPath)
	if uploadRoot == "" {
		return nil, "", "", -1, "上传失败，请稍后再试", nil
	}

	relative := uploadFileName(uid, s.now().Unix(), ext)
	image := "ai_undress/" + relative
	localPath := filepath.Join(uploadRoot, filepath.FromSlash(image))
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return nil, "", "", -1, "上传失败，请稍后再试", err
	}
	out, err := os.Create(localPath)
	if err != nil {
		return nil, "", "", -1, "上传失败，请稍后再试", err
	}
	defer out.Close()
	written, err := io.Copy(out, file)
	if err != nil {
		return nil, "", "", -1, "上传失败，请稍后再试", err
	}
	if written == 0 {
		_ = os.Remove(localPath)
		return nil, "", "", -1, "请选择上传图片", nil
	}

	fileInfo := map[string]interface{}{
		"filename": header.Filename,
		"filesize": int64(float64(written)/1024 + 0.5),
		"suffix":   ext,
		"ispic":    imageSuffix(ext),
		"uri":      relative,
	}
	return fileInfo, localPath, image, 0, "", nil
}

func uploadFileName(uid int, now int64, ext string) string {
	sum := md5.Sum([]byte(fmt.Sprintf("%d%d", uid, now)))
	return hex.EncodeToString(sum[:]) + "." + ext
}

func allowedUploadSuffix(ext string) bool {
	switch ext {
	case "jpg", "jpeg", "gif", "png":
		return true
	default:
		return false
	}
}

func imageSuffix(ext string) int {
	if allowedUploadSuffix(ext) {
		return 1
	}
	return 0
}

func (s *Service) ModuleList(ctx context.Context) (domain.AIUndressExternalData, int, string, error) {
	return s.externalRequest(ctx, "/cps/getModuleList", map[string]interface{}{})
}

func (s *Service) ResourceTypeList(ctx context.Context, module string) (domain.AIUndressExternalData, int, string, error) {
	return s.externalRequest(ctx, "/cps/resourceTypeList", map[string]interface{}{
		"module": module,
	})
}

func (s *Service) ResourceList(ctx context.Context, input domain.AIUndressResourceListInput) (domain.AIUndressExternalData, int, string, error) {
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	return s.externalRequest(ctx, "/cps/resourceList", map[string]interface{}{
		"module":    input.Module,
		"typeId":    input.TypeID,
		"pageSize":  pageSize,
		"current":   input.Current,
		"id":        "",
		"name":      "",
		"sortField": "",
		"sortType":  "",
	})
}

func (s *Service) externalRequest(ctx context.Context, path string, payload map[string]interface{}) (domain.AIUndressExternalData, int, string, error) {
	if s.externalClient == nil {
		return domain.AIUndressExternalData{}, -1, "请求失败", nil
	}
	result, err := s.externalClient.PostJSON(ctx, path, payload)
	if err != nil {
		return domain.AIUndressExternalData{}, -1, "请求失败", nil
	}
	if result.Code != 200 {
		return domain.AIUndressExternalData{}, -1, fmt.Sprintf("请求失败[%d]:%s", result.Code, result.Message), nil
	}
	return domain.AIUndressExternalData{Data: result.Data}, 0, "", nil
}

func (s *Service) userByToken(ctx context.Context, token string) (map[string]interface{}, error) {
	sid := userRepo.CleanToken(token)
	if sid == "" || s.auth == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	user, err := s.auth.UserBySession(ctx, sid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return map[string]interface{}{"uid": "0"}, nil
	}
	return user, nil
}

func (s *Service) aiResourceBaseURL(ctx context.Context) (string, error) {
	if s.env == "test" {
		return "https://pub-21fd0f8233a7476797cc1786f4cabea9.r2.dev", nil
	}
	if s.store == nil {
		return s.resourceBaseURL, nil
	}
	raw, err := s.store.SettingByUUID(ctx, "setting")
	if err != nil {
		return "", err
	}
	value := serializedString(raw, "resurl_h5_ai")
	if value == "" {
		return s.resourceBaseURL, nil
	}
	now := s.now()
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		now = now.In(loc)
	}
	value = strings.ReplaceAll(value, "{rand}", now.Format("2006010215"))
	return strings.TrimRight(value, "/"), nil
}

func (s *Service) resourceURL(baseURL string, value interface{}) interface{} {
	path := strings.TrimSpace(fmt.Sprint(value))
	if path == "" || path == "<nil>" {
		return path
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return path
	}
	return baseURL + "/" + strings.TrimLeft(path, "/")
}

func serializedString(raw string, key string) string {
	pattern := regexp.MustCompile(`s:\d+:"` + regexp.QuoteMeta(key) + `";s:\d+:"([^"]*)"`)
	match := pattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func atoi(value interface{}) int {
	var n int
	_, _ = fmt.Sscan(fmt.Sprint(value), &n)
	return n
}
