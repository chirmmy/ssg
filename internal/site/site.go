package site

import (
	"sort"
	"time"

	"github.com/chirmmy/ssg/internal/config"
	"github.com/chirmmy/ssg/internal/content"
)

type Site struct {
	Config *config.Config
	Posts  []*content.Content
	Pages  map[string]*content.Content
}

// TagCount 标签及其出现次数（列表页的标签云与筛选 chips）。
type TagCount struct {
	Name  string
	Count int
}

// MonthGroup 月份归档：Key 形如 2025-08，Label 形如「2025 年 8 月」，Posts 为该月文章。
type MonthGroup struct {
	Key   string
	Label string
	Count int
	Posts []*content.Content
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

// PostBySlug 供 postLinkCard 短代码解析标题。
func (s *Site) PostBySlug(slug string) *content.Content {
	for _, p := range s.Posts {
		if p.Slug == slug {
			return p
		}
	}
	return nil
}

// TotalChars 全部文章的正文字符数合计（列表页页头统计）。
func (s *Site) TotalChars() int {
	n := 0
	for _, p := range s.Posts {
		n += p.WordCount
	}
	return n
}

// LastUpdated 最新一篇文章的发布时间（Posts 已按时间倒序）。
func (s *Site) LastUpdated() time.Time {
	if len(s.Posts) == 0 {
		return time.Time{}
	}
	return s.Posts[0].PubDate
}

// Featured 精选：优先取 front matter 标了 featured: true 的最新一篇，
// 没有标记时回退到最新一篇。
func (s *Site) Featured() *content.Content {
	if len(s.Posts) == 0 {
		return nil
	}
	for _, p := range s.Posts {
		if p.Featured {
			return p
		}
	}
	return s.Posts[0]
}

// Neighbors 返回相邻文章：prev 更旧、next 更新（Posts 按时间倒序）。
func (s *Site) Neighbors(target *content.Content) (prev, next *content.Content) {
	if target == nil {
		return nil, nil
	}
	for i, p := range s.Posts {
		if p.Slug != target.Slug {
			continue
		}
		if i+1 < len(s.Posts) {
			prev = s.Posts[i+1]
		}
		if i > 0 {
			next = s.Posts[i-1]
		}
		return prev, next
	}
	return nil, nil
}

// Tags 标签聚合：按出现次数降序，次数相同时按首次出现的顺序（保持内容里的直觉顺序）。
func (s *Site) Tags() []TagCount {
	counts := map[string]int{}
	var order []string
	for _, p := range s.Posts {
		for _, t := range p.Tags {
			if _, ok := counts[t]; !ok {
				order = append(order, t)
			}
			counts[t]++
		}
	}
	out := make([]TagCount, 0, len(order))
	for _, t := range order {
		out = append(out, TagCount{Name: t, Count: counts[t]})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}

// Months 月份归档：按时间倒序（Posts 本身已按时间倒序，首次出现即最新）。
func (s *Site) Months() []MonthGroup {
	index := map[string]int{}
	var out []MonthGroup
	for _, p := range s.Posts {
		if p.PubDate.IsZero() {
			continue
		}
		key := p.PubDate.Format("2006-01")
		i, ok := index[key]
		if !ok {
			out = append(out, MonthGroup{
				Key:   key,
				Label: p.PubDate.Format("2006 年 1 月"),
			})
			i = len(out) - 1
			index[key] = i
		}
		out[i].Count++
		out[i].Posts = append(out[i].Posts, p)
	}
	return out
}
