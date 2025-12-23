package crawler

import (
	"social-data-tool/model"
	"time"
)

// GetFallbackData 获取后备数据（基于官方来源手动整理）
// 数据来源：各城市人力资源和社会保障局官方网站
// 更新时间：2024-2025年度
func GetFallbackData() []model.CitySSData {
	year := time.Now().Year()
	now := time.Now().Format(time.RFC3339)

	return []model.CitySSData{
		// 直辖市
		{
			CityCode: "beijing", CityName: "北京", Year: year,
			SSBaseMin: 6326, SSBaseMax: 35283,
			HFBaseMin: 2500, HFBaseMax: 35283,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.28,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.12, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rsj.beijing.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "shanghai", CityName: "上海", Year: year,
			SSBaseMin: 7310, SSBaseMax: 36549,
			HFBaseMin: 2590, HFBaseMax: 36549,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.28,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.07, DefaultHFRateCompany: 0.07,
			SourceURL: "http://rsj.sh.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "tianjin", CityName: "天津", Year: year,
			SSBaseMin: 4400, SSBaseMax: 25983,
			HFBaseMin: 2320, HFBaseMax: 32000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.275,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.11, DefaultHFRateCompany: 0.11,
			SourceURL: "http://hrss.tj.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "chongqing", CityName: "重庆", Year: year,
			SSBaseMin: 4118, SSBaseMax: 22077,
			HFBaseMin: 2100, HFBaseMax: 30000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.275,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.12, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rlsbj.cq.gov.cn/", CrawledAt: now,
		},
		// 广东省
		{
			CityCode: "guangzhou", CityName: "广州", Year: year,
			SSBaseMin: 5284, SSBaseMax: 26421,
			HFBaseMin: 2300, HFBaseMax: 39377,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.27,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.05, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rsj.gz.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "shenzhen", CityName: "深圳", Year: year,
			SSBaseMin: 2360, SSBaseMax: 26421,
			HFBaseMin: 2360, HFBaseMax: 41190,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.265,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.05, DefaultHFRateCompany: 0.05,
			SourceURL: "http://hrss.sz.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "dongguan", CityName: "东莞", Year: year,
			SSBaseMin: 2360, SSBaseMax: 26421,
			HFBaseMin: 2200, HFBaseMax: 35000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.265,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.05, DefaultHFRateCompany: 0.05,
			SourceURL: "http://dghrss.dg.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "foshan", CityName: "佛山", Year: year,
			SSBaseMin: 4546, SSBaseMax: 26421,
			HFBaseMin: 2100, HFBaseMax: 35000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.27,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.05, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rsj.foshan.gov.cn/", CrawledAt: now,
		},
		// 江苏省
		{
			CityCode: "nanjing", CityName: "南京", Year: year,
			SSBaseMin: 4494, SSBaseMax: 24396,
			HFBaseMin: 2280, HFBaseMax: 36000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.nanjing.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "suzhou", CityName: "苏州", Year: year,
			SSBaseMin: 4494, SSBaseMax: 24396,
			HFBaseMin: 2280, HFBaseMax: 36000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.suzhou.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "wuxi", CityName: "无锡", Year: year,
			SSBaseMin: 4494, SSBaseMax: 24396,
			HFBaseMin: 2280, HFBaseMax: 36000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.wuxi.gov.cn/", CrawledAt: now,
		},
		// 浙江省
		{
			CityCode: "hangzhou", CityName: "杭州", Year: year,
			SSBaseMin: 4812, SSBaseMax: 24060,
			HFBaseMin: 2490, HFBaseMax: 39377,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.29,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.12, DefaultHFRateCompany: 0.12,
			SourceURL: "http://hrss.hangzhou.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "ningbo", CityName: "宁波", Year: year,
			SSBaseMin: 4462, SSBaseMax: 22311,
			HFBaseMin: 2280, HFBaseMax: 35000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.29,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.12, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rsj.ningbo.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "wenzhou", CityName: "温州", Year: year,
			SSBaseMin: 4462, SSBaseMax: 22311,
			HFBaseMin: 2280, HFBaseMax: 32000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.29,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.10, DefaultHFRateCompany: 0.10,
			SourceURL: "http://rsj.wenzhou.gov.cn/", CrawledAt: now,
		},
		// 四川省
		{
			CityCode: "chengdu", CityName: "成都", Year: year,
			SSBaseMin: 4071, SSBaseMax: 22555,
			HFBaseMin: 2280, HFBaseMax: 33000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.275,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.06, DefaultHFRateCompany: 0.06,
			SourceURL: "http://cdhrss.chengdu.gov.cn/", CrawledAt: now,
		},
		// 湖北省
		{
			CityCode: "wuhan", CityName: "武汉", Year: year,
			SSBaseMin: 4077, SSBaseMax: 23274,
			HFBaseMin: 2010, HFBaseMax: 30000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.28,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.wuhan.gov.cn/", CrawledAt: now,
		},
		// 湖南省
		{
			CityCode: "changsha", CityName: "长沙", Year: year,
			SSBaseMin: 4000, SSBaseMax: 22491,
			HFBaseMin: 2000, HFBaseMax: 30000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.changsha.gov.cn/", CrawledAt: now,
		},
		// 陕西省
		{
			CityCode: "xian", CityName: "西安", Year: year,
			SSBaseMin: 4559, SSBaseMax: 22795,
			HFBaseMin: 2200, HFBaseMax: 32000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.10, DefaultHFRateCompany: 0.10,
			SourceURL: "http://rsj.xa.gov.cn/", CrawledAt: now,
		},
		// 河南省
		{
			CityCode: "zhengzhou", CityName: "郑州", Year: year,
			SSBaseMin: 3756, SSBaseMax: 20235,
			HFBaseMin: 2000, HFBaseMax: 28000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.zhengzhou.gov.cn/", CrawledAt: now,
		},
		// 山东省
		{
			CityCode: "jinan", CityName: "济南", Year: year,
			SSBaseMin: 4378, SSBaseMax: 23898,
			HFBaseMin: 2200, HFBaseMax: 32000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.12, DefaultHFRateCompany: 0.12,
			SourceURL: "http://jnsi.jinan.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "qingdao", CityName: "青岛", Year: year,
			SSBaseMin: 4378, SSBaseMax: 23898,
			HFBaseMin: 2200, HFBaseMax: 32000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.12, DefaultHFRateCompany: 0.12,
			SourceURL: "http://hrss.qingdao.gov.cn/", CrawledAt: now,
		},
		// 辽宁省
		{
			CityCode: "shenyang", CityName: "沈阳", Year: year,
			SSBaseMin: 4106, SSBaseMax: 20529,
			HFBaseMin: 2100, HFBaseMax: 28000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.275,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.10, DefaultHFRateCompany: 0.10,
			SourceURL: "http://rsj.shenyang.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "dalian", CityName: "大连", Year: year,
			SSBaseMin: 4106, SSBaseMax: 20529,
			HFBaseMin: 2100, HFBaseMax: 28000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.275,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.10, DefaultHFRateCompany: 0.10,
			SourceURL: "http://rsj.dl.gov.cn/", CrawledAt: now,
		},
		// 福建省
		{
			CityCode: "fuzhou", CityName: "福州", Year: year,
			SSBaseMin: 4043, SSBaseMax: 23118,
			HFBaseMin: 2100, HFBaseMax: 30000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.12, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rsj.fuzhou.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "xiamen", CityName: "厦门", Year: year,
			SSBaseMin: 4043, SSBaseMax: 23118,
			HFBaseMin: 2100, HFBaseMax: 32000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.12, DefaultHFRateCompany: 0.12,
			SourceURL: "http://hrss.xm.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "quanzhou", CityName: "泉州", Year: year,
			SSBaseMin: 4043, SSBaseMax: 23118,
			HFBaseMin: 2100, HFBaseMax: 28000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.10, DefaultHFRateCompany: 0.10,
			SourceURL: "http://rsj.quanzhou.gov.cn/", CrawledAt: now,
		},
		// 黑龙江省
		{
			CityCode: "harbin", CityName: "哈尔滨", Year: year,
			SSBaseMin: 3875, SSBaseMax: 19377,
			HFBaseMin: 2000, HFBaseMax: 26000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.27,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.harbin.gov.cn/", CrawledAt: now,
		},
		// 吉林省
		{
			CityCode: "changchun", CityName: "长春", Year: year,
			SSBaseMin: 4071, SSBaseMax: 20355,
			HFBaseMin: 2100, HFBaseMax: 26000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.27,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.changchun.gov.cn/", CrawledAt: now,
		},
		// 安徽省
		{
			CityCode: "hefei", CityName: "合肥", Year: year,
			SSBaseMin: 4071, SSBaseMax: 22023,
			HFBaseMin: 2100, HFBaseMax: 30000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.hefei.gov.cn/", CrawledAt: now,
		},
		// 江西省
		{
			CityCode: "nanchang", CityName: "南昌", Year: year,
			SSBaseMin: 3990, SSBaseMax: 19950,
			HFBaseMin: 2000, HFBaseMax: 26000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.28,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.nc.gov.cn/", CrawledAt: now,
		},
		// 河北省
		{
			CityCode: "shijiazhuang", CityName: "石家庄", Year: year,
			SSBaseMin: 4071, SSBaseMax: 22023,
			HFBaseMin: 2100, HFBaseMax: 28000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.sjz.gov.cn/", CrawledAt: now,
		},
		// 山西省
		{
			CityCode: "taiyuan", CityName: "太原", Year: year,
			SSBaseMin: 3904, SSBaseMax: 19519,
			HFBaseMin: 2000, HFBaseMax: 26000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.28,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.taiyuan.gov.cn/", CrawledAt: now,
		},
		// 云南省
		{
			CityCode: "kunming", CityName: "昆明", Year: year,
			SSBaseMin: 4071, SSBaseMax: 20355,
			HFBaseMin: 2100, HFBaseMax: 28000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.275,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.km.gov.cn/", CrawledAt: now,
		},
		// 贵州省
		{
			CityCode: "guiyang", CityName: "贵阳", Year: year,
			SSBaseMin: 3904, SSBaseMax: 19519,
			HFBaseMin: 2000, HFBaseMax: 26000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.275,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.guiyang.gov.cn/", CrawledAt: now,
		},
		// 广西
		{
			CityCode: "nanning", CityName: "南宁", Year: year,
			SSBaseMin: 3904, SSBaseMax: 19519,
			HFBaseMin: 2000, HFBaseMax: 26000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.275,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.nanning.gov.cn/", CrawledAt: now,
		},
		// 甘肃省
		{
			CityCode: "lanzhou", CityName: "兰州", Year: year,
			SSBaseMin: 3800, SSBaseMax: 19000,
			HFBaseMin: 2000, HFBaseMax: 24000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.27,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.lanzhou.gov.cn/", CrawledAt: now,
		},
		// 江苏省其他城市
		{
			CityCode: "changzhou", CityName: "常州", Year: year,
			SSBaseMin: 4494, SSBaseMax: 24396,
			HFBaseMin: 2280, HFBaseMax: 36000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.285,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.08, DefaultHFRateCompany: 0.08,
			SourceURL: "http://rsj.changzhou.gov.cn/", CrawledAt: now,
		},
		// 广东省其他城市
		{
			CityCode: "zhuhai", CityName: "珠海", Year: year,
			SSBaseMin: 4546, SSBaseMax: 26421,
			HFBaseMin: 2100, HFBaseMax: 35000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.27,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.05, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rsj.zhuhai.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "huizhou", CityName: "惠州", Year: year,
			SSBaseMin: 3376, SSBaseMax: 26421,
			HFBaseMin: 2100, HFBaseMax: 32000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.27,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.05, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rsj.huizhou.gov.cn/", CrawledAt: now,
		},
		{
			CityCode: "zhongshan", CityName: "中山", Year: year,
			SSBaseMin: 4546, SSBaseMax: 26421,
			HFBaseMin: 2100, HFBaseMax: 32000,
			SSPersonRateSum: 0.105, SSCompanyRateSum: 0.27,
			HFRateMin: 0.05, HFRateMax: 0.12,
			DefaultHFRatePerson: 0.05, DefaultHFRateCompany: 0.12,
			SourceURL: "http://rsj.zs.gov.cn/", CrawledAt: now,
		},
	}
}
