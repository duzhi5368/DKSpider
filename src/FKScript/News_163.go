package FKScript

import (
	"github.com/PuerkitoBio/goquery"
	"regexp"
	"strings"
	"FKSpider"
	"FKRequest"
)

func init() {
	News_163_Spider.RegisterToSpiderSpecies()
}

var News_163_Spider = &FKSpider.Spider{
	Name:        "新闻站_网易新闻",
	Description: "网易排行榜新闻，抓取新闻内容 [news.163.com/rank]",
	EnableCookie: false,
	RuleTree: &FKSpider.RuleTree{
		Root: func(ctx *FKSpider.Context) {
			ctx.AddQueue(&FKRequest.Request{Url: "http://news.163.com/rank/", Rule: "排行榜主页"})
		},

		Trunk: map[string]*FKSpider.Rule{

			"排行榜主页": {
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()
					query.Find(".subNav a").Each(func(i int, s *goquery.Selection) {
						if url, ok := s.Attr("href"); ok {
							ctx.AddQueue(&FKRequest.Request{Url: url, Rule: "新闻排行榜"})
						}
					})
				},
			},

			"新闻排行榜": {
				ParseFunc: func(ctx *FKSpider.Context) {
					topTit := []string{
						"1小时前点击排行",
						"24小时点击排行",
						"本周点击排行",
						"今日跟帖排行",
						"本周跟帖排行",
						"本月跟贴排行",
					}
					query := ctx.GetDom()
					// 获取新闻分类
					newsType := query.Find(".titleBar h2").Text()

					urls_top := map[string]string{}

					query.Find(".tabContents").Each(func(n int, t *goquery.Selection) {
						t.Find("tr").Each(func(i int, s *goquery.Selection) {
							// 跳过标题栏
							if i == 0 {
								return
							}
							// 内容链接
							url, ok := s.Find("a").Attr("href")

							// 排名
							top := s.Find(".cBlue").Text()

							if ok {
								urls_top[url] += topTit[n] + ":" + top + ","
							}
						})
					})
					for k, v := range urls_top {
						ctx.AddQueue(&FKRequest.Request{
							Url:  k,
							Rule: "热点新闻",
							Temp: map[string]interface{}{
								"newsType": newsType,
								"top":      v,
							},
						})
					}
				},
			},

			"热点新闻": {
				//注意：有无字段语义和是否输出数据必须保持一致
				ItemFields: []string{
					"标题",
					"内容",
					"排名",
					"类别",
					"ReleaseTime",
				},
				ParseFunc: func(ctx *FKSpider.Context) {
					query := ctx.GetDom()

					// 若有多页内容，则获取阅读全文的链接并获取内容
					if pageAll := query.Find(".ep-pages-all"); len(pageAll.Nodes) != 0 {
						if pageAllUrl, ok := pageAll.Attr("href"); ok {
							ctx.AddQueue(&FKRequest.Request{
								Url:  pageAllUrl,
								Rule: "热点新闻",
								Temp: ctx.CopyTemps(),
							})
						}
						return
					}

					// 获取标题
					title := query.Find("#h1title").Text()

					// 获取内容
					content := query.Find("#endText").Text()
					re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
					content = re.ReplaceAllString(content, "")

					// 获取发布日期
					release := query.Find(".ep-time-soure").Text()
					release = strings.Split(release, "来源:")[0]
					release = strings.Trim(release, " \t\n")

					// 结果存入Response中转
					ctx.Output(map[int]interface{}{
						0: title,
						1: content,
						2: ctx.GetTemp("top", ""),
						3: ctx.GetTemp("newsType", ""),
						4: release,
					})
				},
			},
		},
	},
}