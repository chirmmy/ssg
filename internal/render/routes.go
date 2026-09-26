package render

import (
	"github.com/chirmmy/ssg/internal/site"
)

type Route struct {
	Pattern  string
	Template string
	Data     interface{}
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
		pattern := "/" + slug + "/"
		if slug == "index" {
			pattern = "/"
		}
		routes = append(routes, Route{
			Pattern:  pattern,
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
