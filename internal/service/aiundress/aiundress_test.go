package aiundress

import (
	"context"
	"errors"
	"reflect"
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
	setting string
	rows    []map[string]interface{}
	row     map[string]interface{}
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

func (f fakeStore) SettingByUUID(context.Context, string) (string, error) {
	return f.setting, nil
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

func TestRequireLoginEdge(t *testing.T) {
	service := NewService(fakeAuth{}, fakeStore{}, "https://res.example")

	retcode, errmsg, err := service.RequireLoginEdge(context.Background(), "", "pending")
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "请先登录" {
		t.Fatalf("retcode=%d errmsg=%q", retcode, errmsg)
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

func TestDeleteEdgeExistingRowStopsBeforeWrite(t *testing.T) {
	service := NewService(
		fakeAuth{user: map[string]interface{}{"uid": "7"}},
		fakeStore{row: map[string]interface{}{"id": "99", "image": "a.jpg", "output": "b.jpg"}},
		"https://res.example",
	)

	retcode, errmsg, err := service.DeleteEdge(context.Background(), "250f790ba71ec2b9d3855f424db2259e", 99)
	if err != nil {
		t.Fatal(err)
	}
	if retcode != -1 || errmsg != "AI 删除成功分支暂未迁移" {
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
