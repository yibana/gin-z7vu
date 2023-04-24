package amazon

type ProductValues struct {
	Price        float64 `json:"price"`
	RatingsCount float64 `json:"ratingsCount"`
	Rating       float64 `json:"rating"`
	ReviewCount  float64 `json:"reviewCount"`
	MainRanking  float64 `json:"mainRanking"`
	SubRanking   float64 `json:"subRanking"`
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
}
