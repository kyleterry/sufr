package app

import "github.com/kyleterry/sufr/pkg/data"

type SiteEnvelope struct {
	Title       string
	Flashes     map[string][]string
	CurrentUser *UserEnvelope
	Settings    *SettingsEnvelope
	Sidebar     *SidebarEnvelope
	Version     string
	BuildCommit string
}

type SettingsEnvelope struct{}

type SidebarEnvelope struct {
	PinnedTags *TagListEnvelope
}

type UserEnvelope struct {
	Email string
}

type URLListEnvelope struct {
	Paginator data.Paginator
}

type URLEnvelope struct{}

type TagListEnvelope struct{}

type TagEnvelope struct{}
