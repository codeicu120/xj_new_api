package aiundress

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

type fakeAuth struct {
	user map[string]interface{}
}

func (f fakeAuth) UserBySession(context.Context, string) (map[string]interface{}, error) {
	return f.user, nil
}

type fakeStore struct {
	setting         string
	rows            []map[string]interface{}
	row             map[string]interface{}
	createdUpload   *uploadRecord
	refreshedUpload *uploadRecord
}

type uploadRecord struct {
	uid   int
	id    int
	image string
	now   int64
}

func (f fakeStore) Count(context.Context, int, int) (int, error) {
	return len(f.rows), nil
}

func (f fakeStore) List(context.Context, int, int, int, int, int) ([]map[string]interface{}, error) {
	return f.rows, nil
}

func (f fakeStore) ByID(context.Context, int) (map[string]interface{}, error) {
	if f.row != nil {
		return f.row, nil
	}
	return map[string]interface{}{}, nil
}

func (f fakeStore) ByUIDImage(context.Context, int, string) (map[string]interface{}, error) {
	if f.row != nil {
		return f.row, nil
	}
	return map[string]interface{}{}, nil
}

func (f fakeStore) CreateUpload(_ context.Context, uid int, image string, now int64) (int, error) {
	if f.createdUpload != nil {
		*f.createdUpload = uploadRecord{uid: uid, image: image, now: now}
	}
	return 99, nil
}

func (f fakeStore) RefreshUpload(_ context.Context, id int, now int64) error {
	if f.refreshedUpload != nil {
		*f.refreshedUpload = uploadRecord{id: id, now: now}
	}
	return nil
}

func (f fakeStore) MarkDeleted(context.Context, int, int64) error {
	return nil
}

func (f fakeStore) SettingByUUID(context.Context, string) (string, error) {
	return f.setting, nil
}

type fakeDeleteStore struct {
	fakeStore
	markedID   int
	updateTime int64
	markErr    error
}

func (f *fakeDeleteStore) MarkDeleted(_ context.Context, id int, updateTime int64) error {
	f.markedID = id
	f.updateTime = updateTime
	return f.markErr
}

type fakeExternalClient struct {
	path    string
	payload map[string]interface{}
	resp    ExternalResponse
	err     error
}

func (c *fakeExternalClient) PostJSON(_ context.Context, path string, payload map[string]interface{}) (ExternalResponse, error) {
	c.path = path
	c.payload = payload
	return c.resp, c.err
}

type fakeUploader struct {
	localPath string
	objectKey string
	err       error
}

func (u *fakeUploader) Upload(_ context.Context, localPath string, objectKey string) error {
	u.localPath = localPath
	u.objectKey = objectKey
	return u.err
}

type memoryMultipartFile struct {
	*bytes.Reader
}

func (memoryMultipartFile) Close() error {
	return nil
}

func newUploadFile(content string, filename string) (multipart.File, *multipart.FileHeader) {
	reader := bytes.NewReader([]byte(content))
	return memoryMultipartFile{Reader: reader}, &multipart.FileHeader{
		Filename: filename,
		Size:     int64(len(content)),
	}
}

func TestListingRequiresLoginWithPHPErrorCode(t *testing.T) {
	service := NewService(fakeAuth{}, fakeStore{}, "https://res.example")

	_, retcode, errmsg, err := service.Listing(context.Background(), "", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请先登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestUploadRequiresLogin(t *testing.T) {
	service := NewService(fakeAuth{}, fakeStore{}, "https://res.example")
	file, header := newUploadFile("image", "face.jpg")

	_, retcode, errmsg, err := service.Upload(context.Background(), "", file, header)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请先登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestUploadCreatesFileAndDBRow(t *testing.T) {
	created := uploadRecord{}
	uploader := &fakeUploader{}
	store := fakeStore{createdUpload: &created}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.example").
		WithUploadPath(t.TempDir()).
		WithObjectUploader(uploader)
	service.now = func() time.Time { return time.Unix(123456, 0) }
	file, header := newUploadFile("image-bytes", "换脸模版.jpg")

	data, retcode, errmsg, err := service.Upload(context.Background(), "250f790ba71ec2b9d3855f424db2259e", file, header)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	fileData := data["file"].(map[string]interface{})
	uri := fileData["uri"].(string)
	if !strings.HasPrefix(uri, "ai_undress/") || !strings.HasSuffix(uri, ".jpg") || strings.Count(uri, "/") != 1 {
		t.Fatalf("unexpected uri %q", uri)
	}
	if fileData["filename"] != "换脸模版.jpg" || fileData["suffix"] != "jpg" || fileData["filesize"] != int64(0) || fileData["ispic"] != 1 {
		t.Fatalf("unexpected file data %#v", fileData)
	}
	if created.uid != 7 || created.image != uri || created.now != 123456 {
		t.Fatalf("created upload=%#v", created)
	}
	if uploader.objectKey != uri {
		t.Fatalf("uploader object key=%q uri=%q", uploader.objectKey, uri)
	}
	if _, err := os.Stat(uploader.localPath); err != nil {
		t.Fatalf("expected local file to exist: %v", err)
	}
}

func TestUploadRefreshesExistingRow(t *testing.T) {
	refreshed := uploadRecord{}
	store := fakeStore{
		row:             map[string]interface{}{"id": "10", "image": "ai_undress/existing.jpg"},
		refreshedUpload: &refreshed,
	}
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, store, "https://res.example").
		WithUploadPath(t.TempDir())
	service.now = func() time.Time { return time.Unix(123456, 0) }
	file, header := newUploadFile("image-bytes", "face.png")

	_, retcode, errmsg, err := service.Upload(context.Background(), "token", file, header)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if refreshed.id != 10 || refreshed.now != 123456 {
		t.Fatalf("refreshed upload=%#v", refreshed)
	}
}

func TestUploadRejectsUploaderFailure(t *testing.T) {
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, fakeStore{}, "https://res.example").
		WithUploadPath(t.TempDir()).
		WithObjectUploader(&fakeUploader{err: errors.New("r2 down")})
	file, header := newUploadFile("image-bytes", "face.jpg")

	_, retcode, errmsg, err := service.Upload(context.Background(), "token", file, header)
	if err == nil {
		t.Fatal("expected uploader error")
	}
	if retcode != -1 || errmsg != "上传失败，请稍后再试" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestUploadRejectsUnsupportedSuffix(t *testing.T) {
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, fakeStore{}, "https://res.example").
		WithUploadPath(t.TempDir())
	file, header := newUploadFile("image-bytes", "face.webp")

	_, retcode, errmsg, err := service.Upload(context.Background(), "token", file, header)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "face.webp:系统只允许上传[jpg, jpeg, gif, png]格式的文件" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestUploadKeepsJPEGSuffix(t *testing.T) {
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, fakeStore{}, "https://res.example").
		WithUploadPath(t.TempDir())
	service.now = func() time.Time { return time.Unix(123456, 0) }
	file, header := newUploadFile("image-bytes", "face.jpeg")

	data, retcode, errmsg, err := service.Upload(context.Background(), "token", file, header)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	fileData := data["file"].(map[string]interface{})
	uri := fileData["uri"].(string)
	if !strings.HasSuffix(uri, ".jpeg") || fileData["suffix"] != "jpeg" || fileData["ispic"] != 1 {
		t.Fatalf("unexpected file data %#v", fileData)
	}
}

func TestUndressEdgeInvalidImagePath(t *testing.T) {
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, fakeStore{}, "https://res.example")

	retcode, errmsg, err := service.UndressEdge(context.Background(), "250f790ba71ec2b9d3855f424db2259e", "ai_undress/missing.jpg", 0)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "无效图片路径" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestUndressEdgeExistingImageStopsBeforeGeneration(t *testing.T) {
	service := NewService(
		fakeAuth{user: map[string]interface{}{"uid": "7"}},
		fakeStore{row: map[string]interface{}{"id": "10", "uid": "7", "image": "ai_undress/a.jpg", "module": "0"}},
		"https://res.example",
	)

	retcode, errmsg, err := service.UndressEdge(context.Background(), "250f790ba71ec2b9d3855f424db2259e", "ai_undress/a.jpg", 9)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "AI 生成成功分支暂未迁移" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestDeleteEdgeMissingRowReturnsOK(t *testing.T) {
	service := NewService(fakeAuth{user: map[string]interface{}{"uid": "7"}}, fakeStore{}, "https://res.example")

	retcode, errmsg, err := service.DeleteEdge(context.Background(), "250f790ba71ec2b9d3855f424db2259e", 99)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestDeleteEdgeExistingRowMarksDeletedAndRemovesFiles(t *testing.T) {
	dir := t.TempDir()
	image := filepath.Join(dir, "ai_undress", "a.jpg")
	output := filepath.Join(dir, "ai_undress", "b.jpg")
	if err := os.MkdirAll(filepath.Dir(image), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(image, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("output"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := &fakeDeleteStore{fakeStore: fakeStore{row: map[string]interface{}{"id": "99", "image": "ai_undress/a.jpg", "output": "ai_undress/b.jpg"}}}
	service := NewService(
		fakeAuth{user: map[string]interface{}{"uid": "7"}},
		store,
		"https://res.example",
	).WithUploadPath(dir)
	service.now = func() time.Time { return time.Unix(12345, 0) }

	retcode, errmsg, err := service.DeleteEdge(context.Background(), "250f790ba71ec2b9d3855f424db2259e", 99)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if store.markedID != 99 || store.updateTime != 12345 {
		t.Fatalf("marked id=%d updateTime=%d", store.markedID, store.updateTime)
	}
	if _, err := os.Stat(image); !os.IsNotExist(err) {
		t.Fatalf("image should be removed, stat err=%v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("output should be removed, stat err=%v", err)
	}
}

func TestDeleteEdgeMarkFailure(t *testing.T) {
	store := &fakeDeleteStore{
		fakeStore: fakeStore{row: map[string]interface{}{"id": "99", "image": "missing.jpg", "output": ""}},
		markErr:   errors.New("db down"),
	}
	service := NewService(
		fakeAuth{user: map[string]interface{}{"uid": "7"}},
		store,
		"https://res.example",
	).WithUploadPath(t.TempDir())

	retcode, errmsg, err := service.DeleteEdge(context.Background(), "250f790ba71ec2b9d3855f424db2259e", 99)
	if err == nil {
		t.Fatal("expected error")
	}
	if retcode != -1 || errmsg != "AI 删除失败" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestListingBuildsResourceURLsAndPageInfo(t *testing.T) {
	service := NewService(
		fakeAuth{user: map[string]interface{}{"uid": "7"}},
		fakeStore{
			setting: `a:1:{s:12:"resurl_h5_ai";s:36:"https://ai-{rand}.example.com/assets";}`,
			rows: []map[string]interface{}{
				{"id": "1", "uid": "7", "image": "ai_undress/a.jpg", "output": "out/b.jpg"},
			},
		},
		"https://res.example",
	)
	service.now = func() time.Time { return time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC) }

	data, retcode, errmsg, err := service.Listing(context.Background(), "250f790ba71ec2b9d3855f424db2259e", 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if data.Rows[0]["image"] != "https://ai-2026071517.example.com/assets/ai_undress/a.jpg" {
		t.Fatalf("image url = %#v", data.Rows[0]["image"])
	}
	if data.Rows[0]["output"] != "https://ai-2026071517.example.com/assets/out/b.jpg" {
		t.Fatalf("output url = %#v", data.Rows[0]["output"])
	}
	if data.PageInfo["page_url"] != "/aiundress/listing?page=[?]" || data.PageInfo["pagesize"] != 10 {
		t.Fatalf("pageinfo = %#v", data.PageInfo)
	}
}

func TestModuleListWrapsExternalData(t *testing.T) {
	client := &fakeExternalClient{resp: ExternalResponse{
		Code: 200,
		Data: []interface{}{map[string]interface{}{"module": "4"}},
	}}
	service := NewService(fakeAuth{}, fakeStore{}, "").WithExternalClient(client)

	data, retcode, errmsg, err := service.ModuleList(context.Background())
	if err != nil {
		t.Fatalf("module list: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if client.path != "/cps/getModuleList" {
		t.Fatalf("path=%q", client.path)
	}
	if len(client.payload) != 0 {
		t.Fatalf("payload=%#v", client.payload)
	}
	if !reflect.DeepEqual(data.Data, client.resp.Data) {
		t.Fatalf("data=%#v", data.Data)
	}
}

func TestResourceTypeListForwardsModule(t *testing.T) {
	client := &fakeExternalClient{resp: ExternalResponse{Code: 200, Data: map[string]interface{}{"ok": true}}}
	service := NewService(fakeAuth{}, fakeStore{}, "").WithExternalClient(client)

	_, retcode, errmsg, err := service.ResourceTypeList(context.Background(), "4")
	if err != nil {
		t.Fatalf("resource type list: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	if client.path != "/cps/resourceTypeList" {
		t.Fatalf("path=%q", client.path)
	}
	if client.payload["module"] != "4" {
		t.Fatalf("payload=%#v", client.payload)
	}
}

func TestResourceListForwardsPHPPayloadAndDefaultPageSize(t *testing.T) {
	client := &fakeExternalClient{resp: ExternalResponse{Code: 200, Data: map[string]interface{}{"rows": []interface{}{}}}}
	service := NewService(fakeAuth{}, fakeStore{}, "").WithExternalClient(client)

	_, retcode, errmsg, err := service.ResourceList(context.Background(), ResourceListInput{
		Module:  "4",
		TypeID:  "12",
		Current: "3",
	})
	if err != nil {
		t.Fatalf("resource list: %v", err)
	}
	if retcode != 0 || errmsg != "" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
	want := map[string]interface{}{
		"module":    "4",
		"typeId":    "12",
		"pageSize":  10,
		"current":   "3",
		"id":        "",
		"name":      "",
		"sortField": "",
		"sortType":  "",
	}
	if client.path != "/cps/resourceList" {
		t.Fatalf("path=%q", client.path)
	}
	if !reflect.DeepEqual(client.payload, want) {
		t.Fatalf("payload=%#v want %#v", client.payload, want)
	}
}

func TestExternalFailureMatchesPHPRequestFailed(t *testing.T) {
	service := NewService(fakeAuth{}, fakeStore{}, "")

	_, retcode, errmsg, err := service.ModuleList(context.Background())
	if err != nil {
		t.Fatalf("module list: %v", err)
	}
	if retcode != -1 || errmsg != "请求失败" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}

	service.WithExternalClient(&fakeExternalClient{err: errors.New("dial failed")})
	_, retcode, errmsg, err = service.ModuleList(context.Background())
	if err != nil {
		t.Fatalf("module list: %v", err)
	}
	if retcode != -1 || errmsg != "请求失败" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}

func TestExternalBusinessFailureMatchesPHPMessage(t *testing.T) {
	service := NewService(fakeAuth{}, fakeStore{}, "").WithExternalClient(&fakeExternalClient{resp: ExternalResponse{
		Code:    401,
		Message: "bad key",
	}})

	_, retcode, errmsg, err := service.ModuleList(context.Background())
	if err != nil {
		t.Fatalf("module list: %v", err)
	}
	if retcode != -1 || errmsg != "请求失败[401]:bad key" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
	}
}
