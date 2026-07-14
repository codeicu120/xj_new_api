package user

import "testing"

func TestSysAvatarServiceList(t *testing.T) {
	service := NewSysAvatarService("https://static.example.test/")

	data := service.List()

	man := data.SysAvatar["man"]
	if len(man) != 9 {
		t.Fatalf("expected 9 man avatars, got %d", len(man))
	}
	if man[0] != "https://static.example.test/sysavatar/man/1.png" {
		t.Fatalf("unexpected first man avatar %q", man[0])
	}

	woman := data.SysAvatar["woman"]
	if len(woman) != 9 {
		t.Fatalf("expected 9 woman avatars, got %d", len(woman))
	}
	if woman[8] != "https://static.example.test/sysavatar/woman/9.png" {
		t.Fatalf("unexpected last woman avatar %q", woman[8])
	}
}
