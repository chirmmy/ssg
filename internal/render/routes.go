package render

import (
	"github.com/chirmmy/ssg/internal/site"
)

type Route struct {
	Pattern  string
	Template string
	Data     interface{}
	OutFile  string // 可选。指定时直接用这个路径（相对 outDir）
}

type HomeData struct {
	Site  *site.Site
	Posts []interface{}
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
		routes = append(routes, Route{
			Pattern:  "/blog/" + p.Slug + "/",
			Template: "post",
			Data: map[string]interface{}{
				"Site": s,
				"Post": p,
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
				Data: map[string]any{
					"Site": s,
					"Page": page,
				},
			})
			continue
		}

		routes = append(routes, Route{
			Pattern:  "/" + slug + "/",
			Template: "page",
			Data: map[string]interface{}{
				"Site": s,
				"Page": page,
			},
		})
	}

	return routes
}

func homeData(s *site.Site) map[string]interface{} {
	return map[string]interface{}{
		"Site":  s,
		"Posts": s.LatestPosts(3),
	}
}

func blogIndexData(s *site.Site) map[string]interface{} {
	return map[string]interface{}{
		"Site":  s,
		"Posts": s.Posts,
	}
}
