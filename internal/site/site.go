package site

import (
	"github.com/chirmmy/ssg/internal/config"
	"github.com/chirmmy/ssg/internal/content"
)

type Site struct {
	Config *config.Config
	Posts  []*content.Content
	Pages  map[string]*content.Content
}

func NewSite(cfg *config.Config, items []*content.Content) *Site {
	site := &Site{
		Config: cfg,
		Pages:  make(map[string]*content.Content),
	}

	for _, c := range items {
		if c.Draft && !cfg.Build.Drafts {
			continue
		}
		switch c.Kind {
		case content.KindPage:
			site.Pages[c.Slug] = c
		case content.KindPost:
			site.Posts = append(site.Posts, c)
		}
	}
	return site
}

func (s *Site) Page(slug string) *content.Content {
	return s.Pages[slug]
}

func (s *Site) LatestPosts(n int) []*content.Content {
	if n > len(s.Posts) {
		n = len(s.Posts)
	}
	return s.Posts[:n]
}
