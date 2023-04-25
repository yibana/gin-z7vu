package amazon

import "regexp"

type ProductValues struct {
	Price        float64 `json:"price"`
	RatingsCount float64 `json:"ratingsCount"`
	Rating       float64 `json:"rating"`
	ReviewCount  float64 `json:"reviewCount"`
	MainRanking  float64 `json:"mainRanking"`
	SubRanking   float64 `json:"subRanking"`
}
type DeliveryInfo struct {
	Mode string
	Info map[string]string `json:"info,omitempty"`
}
type Product struct {
	ASIN          string              `json:"asin"`                   // 产品编号
	Title         string              `json:"title"`                  // 产品标题
	Price         string              `json:"price"`                  // 产品价格
	Brand         string              `json:"brand,omitempty"`        // 产品品牌
	MerchantInfo  string              `json:"merchantInfo,omitempty"` // 产品商家信息
	RatingsCount  string              `json:"ratingsCount,omitempty"` // 产品评价数量
	ListingDate   string              `json:"listingDate,omitempty"`  // 上架日期
	Size          string              `json:"size,omitempty"`         // 产品尺寸
	Rating        string              `json:"rating,omitempty"`       // 产品评分
	ReviewCount   string              `json:"reviewCount,omitempty"`  // 产品评价数量
	MainRanking   string              `json:"mainRanking,omitempty"`  // 产品主排名
	SubRanking    string              `json:"subRanking,omitempty"`   // 产品子排名
	Details       []map[string]string `json:"details,omitempty"`      // 产品细节
	ProductValues ProductValues       `json:"productValues,omitempty"`
	DeliveryInfo  DeliveryInfo        `json:"deliveryInfo,omitempty"`
}

func MerchantInfo2DeliveryInfo(MerchantInfo string) (dinfo DeliveryInfo) {
	/*
		https://www.amazon.ca/gp/product/B09F6J9J3M?th=1&psc=1
		Ships from China and sold by zhanghuawei.
		FBM发货方式（卖家自己发货），发货地是China。卖家名称zhanghuawei

		https://www.amazon.ca/gp/product/B09X1CLFMN?th=1&psc=1
		Sold by HNFA【CA】 and Fulfilled by Amazon.
		FBA发货模式（亚马逊仓库发货），卖家名称是HNFA【CA】
		Sold by zijiepaddle and Fulfilled by Amazon from outside Canada. Customs & Duties may apply. Importers of commercial goods should review the shipping & delivery policy.
		这个也是FBA模式，只是说仓库可能不在加拿大。美国/加拿大/墨西哥，公用一份数据。卖家名称是zijiepaddle

		https://www.amazon.ca/gp/product/B00AY80XR8?th=1
		Ships from and sold by Amazon.ca.
		AMZ自营商品。
	*/
	dinfo.Info = make(map[string]string)
	if MerchantInfo == "" {
		return
	}
	if MerchantInfo == "Ships from and sold by Amazon.ca." {
		dinfo.Mode = "AMZ"
		dinfo.Info["sellerName"] = "Amazon.ca"
		dinfo.Info["shipFrom"] = "Amazon.ca"
		return
	}
	// 用正则提取卖家名称
	compile := regexp.MustCompile(`Ships from\s+(.*?)\s+and sold by\s+(.*?)\.`)
	submatch := compile.FindStringSubmatch(MerchantInfo)
	if len(submatch) == 3 {
		dinfo.Mode = "FBM"
		dinfo.Info["shipFrom"] = submatch[1]
		dinfo.Info["sellerName"] = submatch[2]
		return
	}
	compile = regexp.MustCompile(`Sold by\s+(.*?)\s+and Fulfilled by\s+(.*?)\.`)
	submatch = compile.FindStringSubmatch(MerchantInfo)
	if len(submatch) == 3 {
		dinfo.Mode = "FBA"
		dinfo.Info["sellerName"] = submatch[1]
		dinfo.Info["fulfilledBy"] = submatch[2]
		return
	}
	compile = regexp.MustCompile(`Sold by\s+(.*?)\s+and Fulfilled by\s+(.*?)\s+from outside\s+(.*?)\.`)
	submatch = compile.FindStringSubmatch(MerchantInfo)
	if len(submatch) == 4 {
		dinfo.Mode = "FBA"
		dinfo.Info["sellerName"] = submatch[1]
		dinfo.Info["fulfilledBy"] = submatch[2]
		dinfo.Info["shipFromOutside"] = submatch[3]
		return
	}
	return
}
