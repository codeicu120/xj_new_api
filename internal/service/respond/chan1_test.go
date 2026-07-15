package respond

import (
	"context"
	"testing"
)

type fakeChan1Store struct {
	seenMobi     string
	user         map[string]interface{}
	paymentCount int
	pkg          map[string]interface{}
}

func (s *fakeChan1Store) UserByMobi(_ context.Context, mobi string) (map[string]interface{}, error) {
	s.seenMobi = mobi
	return s.user, nil
}

func (s *fakeChan1Store) CountPaymentsByUIDPayTypePayway(context.Context, int, int, string) (int, error) {
	return s.paymentCount, nil
}

func (s *fakeChan1Store) VIPPackageByID(context.Context, int) (map[string]interface{}, error) {
	return s.pkg, nil
}

func TestChan1NormalizesDigitsAndRejectsMissingUser(t *testing.T) {
	store := &fakeChan1Store{}
	service := NewService(store)

	retcode, errmsg, err := service.Chan1(context.Background(), "18812345678")
	if err != nil {
		t.Fatalf("chan1: %v", err)
	}
	if store.seenMobi != "86.18812345678" {
		t.Fatalf("mobi = %q", store.seenMobi)
	}
	if retcode != 2 || errmsg != "用户不存在" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestChan1RejectsAlreadyGiftedUser(t *testing.T) {
	service := NewService(&fakeChan1Store{
		user:         map[string]interface{}{"uid": "8"},
		paymentCount: 1,
	})

	retcode, errmsg, err := service.Chan1(context.Background(), "86.18812345678")
	if err != nil {
		t.Fatalf("chan1: %v", err)
	}
	if retcode != 3 || errmsg != "该用户已经送过会员了" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}

func TestChan1RejectsMissingOrStoppedPackage(t *testing.T) {
	for _, pkg := range []map[string]interface{}{
		{},
		{"pkgid": "1", "showtype": "1"},
	} {
		service := NewService(&fakeChan1Store{
			user: map[string]interface{}{"uid": "8"},
			pkg:  pkg,
		})

		retcode, errmsg, err := service.Chan1(context.Background(), "86.18812345678")
		if err != nil {
			t.Fatalf("chan1: %v", err)
		}
		if retcode != -1 || errmsg != "套餐不存在或未启用" {
			t.Fatalf("unexpected response %d %q", retcode, errmsg)
		}
	}
}

func TestChan1SuccessBranchStillPending(t *testing.T) {
	service := NewService(&fakeChan1Store{
		user: map[string]interface{}{"uid": "8"},
		pkg:  map[string]interface{}{"pkgid": "1", "showtype": "0"},
	})

	retcode, errmsg, err := service.Chan1(context.Background(), "86.18812345678")
	if err != nil {
		t.Fatalf("chan1: %v", err)
	}
	if retcode != -1 || errmsg != "chan1 成功分支暂未迁移" {
		t.Fatalf("unexpected response %d %q", retcode, errmsg)
	}
}
