package render

import (
	"time"

	"github.com/chirmmy/ssg/internal/content"
	"github.com/chirmmy/ssg/internal/site"
)

type Route struct {
	Pattern  string
	Template string
	Data     interface{}
	OutFile  string // 可选。指定时直接用这个路径（相对 outDir）
}

// pinnedTagCount 列表页筛选栏里常驻显示的标签个数，其余收进「更多」。
const pinnedTagCount = 8

// PageMeta 是所有页面数据共有的字段，便于 head.html / header.html 统一取用。
type PageMeta struct {
	Title        string // <title>
	Current      string // 顶栏当前栏目：blog / about / 空
	ShowProgress bool   // 顶栏下沿的阅读进度条（仅文章页）
}

type HomeData struct {
	PageMeta
	Site  *site.Site
	Posts []*content.Content
}

type PostData struct {
	PageMeta
	Site *site.Site
	Post *content.Content
	Prev *content.Content // 上一篇文章（更旧）
	Next *content.Content // 下一篇文章（更新）
}

type BlogIndexData struct {
	PageMeta
	Site        *site.Site
	Posts       []*content.Content
	Featured    *content.Content
	Tags        []site.TagCount
	PinnedTags  []site.TagCount
	RestTags    []site.TagCount
	Months      []site.MonthGroup
	Total       int
	TotalChars  int
	LastUpdated time.Time
}

type PageData struct {
	PageMeta
	Site *site.Site
	Page *content.Content
}

func BuildRoutes(s *site.Site) []Route {
	routes := []Route{
		{
			Pattern:  "/",
			Template: "home",
			Data:     homeData(s),
		},
		{
			Pattern:  "/blog/",
			Template: "blog-index",
			Data:     blogIndexData(s),
		},
	}

	for _, p := range s.Posts {
		prev, next := s.Neighbors(p)
		routes = append(routes, Route{
			Pattern:  "/blog/" + p.Slug + "/",
			Template: "post",
			Data: PostData{
				PageMeta: PageMeta{
					Title:        p.Title, // 供 head.html 的 <title> 使用
					Current:      "blog",
					ShowProgress: true,
				},
				Site: s,
				Post: p,
				Prev: prev,
				Next: next,
			},
		})
	}
	for slug, page := range s.Pages {
		switch slug {
		case "index": // index.md 不作为独立页面输出，首页由 home 模板负责
			continue
		case "404":
			routes = append(routes, Route{
				Pattern:  "/404/",
				Template: "404",
				OutFile:  "404.html",
				Data: PageData{
					PageMeta: PageMeta{Title: page.Title},
					Site:     s,
					Page:     page,
				},
			})
			continue
		}

		current := ""
		if slug == "about" {
			current = "about"
		}
		routes = append(routes, Route{
			Pattern:  "/" + slug + "/",
			Template: "page",
			Data: PageData{
				PageMeta: PageMeta{
					Title:   page.Title, // 供 head.html 的 <title> 使用
					Current: current,
				},
				Site: s,
				Page: page,
			},
		})
	}

	return routes
}

func homeData(s *site.Site) HomeData {
	return HomeData{
		Site:  s,
		Posts: s.LatestPosts(3),
	}
}

func blogIndexData(s *site.Site) BlogIndexData {
	tags := s.Tags()
	pinned := tags
	rest := []site.TagCount{}
	if len(pinned) > pinnedTagCount {
		pinned = tags[:pinnedTagCount]
		rest = tags[pinnedTagCount:]
	}
	return BlogIndexData{
		PageMeta: PageMeta{
			Title:   "博客",
			Current: "blog",
		},
		Site:        s,
		Posts:       s.Posts,
		Featured:    s.Featured(),
		Tags:        tags,
		PinnedTags:  pinned,
		RestTags:    rest,
		Months:      s.Months(),
		Total:       len(s.Posts),
		TotalChars:  s.TotalChars(),
		LastUpdated: s.LastUpdated(),
	}
}
