package user

import (
	"strings"

	"xj_comp/internal/domain"
)

type SysAvatarService struct {
	resourceBaseURL string
}

func NewSysAvatarService(resourceBaseURL string) *SysAvatarService {
	return &SysAvatarService{resourceBaseURL: strings.TrimRight(resourceBaseURL, "/")}
}

func (s *SysAvatarService) List() domain.SysAvatarData {
	avatars := make(map[string][]string, len(systemAvatarGroups))
	for group, paths := range systemAvatarGroups {
		items := make([]string, 0, len(paths))
		for _, path := range paths {
			items = append(items, s.resourceBaseURL+"/"+path)
		}
		avatars[group] = items
	}
	return domain.SysAvatarData{SysAvatar: avatars}
}

var systemAvatarGroups = map[string][]string{
	"man": {
		"sysavatar/man/1.png",
		"sysavatar/man/2.png",
		"sysavatar/man/3.png",
		"sysavatar/man/4.png",
		"sysavatar/man/5.png",
		"sysavatar/man/6.png",
		"sysavatar/man/7.png",
		"sysavatar/man/8.png",
		"sysavatar/man/9.png",
	},
	"woman": {
		"sysavatar/woman/1.png",
		"sysavatar/woman/2.png",
		"sysavatar/woman/3.png",
		"sysavatar/woman/4.png",
		"sysavatar/woman/5.png",
		"sysavatar/woman/6.png",
		"sysavatar/woman/7.png",
		"sysavatar/woman/8.png",
		"sysavatar/woman/9.png",
	},
}
