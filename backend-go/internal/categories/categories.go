package categories

import "strings"

type Category struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

// IDs are permanent. Alias resolution never changes source records.
var All = []Category{
	{"food", "餐饮外卖", []string{"餐饮", "餐饮外卖"}},
	{"transport", "交通出行", []string{"交通", "交通出行"}},
	{"shopping", "综合购物", []string{"购物", "综合购物"}},
	{"groceries", "食品日用", []string{"食品日用"}},
	{"entertainment", "休闲娱乐", []string{"娱乐", "休闲娱乐"}},
	{"housing", "住房水电", []string{"住房", "水电", "utilities", "住房水电"}},
	{"medical", "医疗健康", []string{"医疗", "医疗健康", "healthcare"}},
	{"education", "教育", []string{"教育"}},
	{"communication", "通信费用", []string{"通讯", "通信费用"}},
	{"clothing", "服饰运动", []string{"服饰", "服饰运动"}},
	{"beauty", "美容", []string{"美容"}},
	{"sports", "运动", []string{"运动"}},
	{"travel", "酒店旅行", []string{"旅行", "酒店旅行"}},
	{"social", "人情", []string{"人情", "gifts"}},
	{"pets", "宠物", []string{"宠物", "pet"}},
	{"software", "软件与云服务", []string{"软件与云服务", "云存储"}},
	{"ai", "AI 订阅", []string{"AI 订阅"}},
	{"electronics", "数码家电", []string{"数码家电"}},
	{"fuel", "汽车加油", []string{"汽车加油"}},
	{"charging", "充电费用", []string{"充电费用"}},
	{"insurance", "保险", []string{"保险"}},
	{"platform", "综合平台待细分", []string{"综合平台待细分"}},
	{"uncategorized", "待分类", []string{"待分类", ""}},
	{"other", "其他", []string{"其他"}},
}

func Resolve(key string) Category {
	key = strings.TrimSpace(key)
	for _, c := range All {
		if key == c.ID {
			return c
		}
		for _, alias := range c.Aliases {
			if key == alias {
				return c
			}
		}
	}
	return Category{ID: key, Name: key, Aliases: []string{}}
}
