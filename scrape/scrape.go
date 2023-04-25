package scrape

import (
	"fmt"
	"gin/amazon"
	"gin/utils"
	"github.com/gocolly/colly"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func GetAmzProduct(host, asin, proxy string) (*amazon.Product, error) {
	productURL := fmt.Sprintf("https://%s/dp/%s?th=1&psc=1", host, asin)
	// Create a new collector
	cy := colly.NewCollector(
		colly.AllowedDomains(host),
	)

	if len(proxy) > 0 {
		// 设置代理IP
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		cy.SetProxyFunc(func(_ *http.Request) (*url.URL, error) {
			return proxyURL, nil
		})
	}

	// Create a product object to store the extracted data
	product := &amazon.Product{}

	// On every product page
	var onError error
	cy.OnHTML("#dp-container", func(e *colly.HTMLElement) {
		body := string(e.Response.Body)
		if strings.Contains(body, "Sorry, we just need to make sure you're not a robot. For best results, please make sure your browser is accepting cookies.") {
			onError = fmt.Errorf("robot")
			return
		}
		//ioutil.WriteFile("product.html", []byte(string(e.Response.Body)), 0644)
		product.ASIN = e.ChildAttr("div[data-asin]", "data-asin")
		product.Title = e.ChildText("#productTitle")
		product.Price = utils.TrimAll(e.DOM.Find("div.a-box-inner span.a-price span").First().Text())
		product.Brand = strings.TrimPrefix(e.ChildText("a#bylineInfo"), "Brand: ")
		product.MerchantInfo = utils.TrimAll(e.DOM.Find("#merchant-info").First().Text())
		product.RatingsCount = e.ChildText("#averageCustomerReviews_feature_div #acrCustomerReviewText")
		product.Rating = e.ChildAttr("#acrPopover", "title")
		product.ReviewCount = e.ChildText("#askATFLink")

		availability := e.DOM.Find("#availability span").First()
		product.Availability = utils.TrimAll(availability.Text())

		// 卖家数量
		var sellerSpan []string
		e.ForEach("#olpLinkWidget_feature_div div.olp-text-box>span", func(i int, element *colly.HTMLElement) {
			v := utils.TrimAll(element.DOM.First().Text())
			if len(v) > 0 {
				sellerSpan = append(sellerSpan, element.Text)
			}
		})
		product.OtherSellersSpan = sellerSpan
		table := getAmzTable(e, "table#productDetails_techSpec_section_1")
		if len(table) > 0 {
			product.Details = append(product.Details, table)
		}

		// 产品图片
		e.ForEach("#leftCol li[data-csa-c-element-type=navigational] img", func(i int, element *colly.HTMLElement) {
			product.Images = append(product.Images, element.Attr("src"))
		})

		// 产品详情
		e.ForEach("#tech table", func(i int, e *colly.HTMLElement) {
			var tb = make(map[string]string)
			e.ForEach("tr", func(i int, e *colly.HTMLElement) {
				td := e.DOM.Find("td")
				if td.Length() >= 2 {
					tb[utils.TrimAll(td.Eq(0).Text())] = utils.TrimAll(td.Eq(1).Text())
				}
			})
			if len(tb) > 0 {
				product.Details = append(product.Details, tb)
			}
		})
		// 产品详情
		tb := make(map[string]string)
		e.ForEach("#detailBullets_feature_div span.a-list-item", func(i int, e *colly.HTMLElement) {
			span := e.DOM.Find("span")
			if span.Length() >= 2 {
				tb[utils.TrimSpan(span.Eq(0).Text())] = utils.TrimSpan(span.Eq(1).Text())
			}
		})
		// 产品详情
		e.ForEach("#detailBulletsWrapper_feature_div>ul", func(i int, e *colly.HTMLElement) {
			li := e.DOM.Find("li").First()
			// 过滤掉td中的style元素
			li.Find("style").Remove()
			li.Find("script").Remove()
			span := li.Find("span.a-text-bold").First()
			key := span.Text()
			span.Remove()
			val := li.Text()
			tb[utils.TrimSpan(key)] = utils.TrimAll(val)
		})

		if len(tb) > 0 {
			product.Details = append(product.Details, tb)
		}

		table = getAmzTable(e, "table#productDetails_detailBullets_sections1")
		if len(table) > 0 {
			product.Details = append(product.Details, table)
		}
		product.MainRanking = e.ChildText("#SalesRank")
		if len(product.Details) > 0 {
			var sizes []string

			for _, detail := range product.Details {
				if size, ok := detail["Size"]; ok {
					sizes = append(sizes, size)
				}
				if size, ok := detail["Product Dimensions"]; ok {
					sizes = append(sizes, size)
				}
				if date, ok := detail["Date First Available"]; ok {
					product.ListingDate = date
				}
				if brand, ok := detail["Brand"]; ok {
					product.Brand = brand
				}
				if ranks, ok := detail["Best Sellers Rank"]; ok {
					rank_arr := strings.Split(ranks, " #")
					if len(rank_arr) >= 2 {
						product.MainRanking = utils.TrimAll(rank_arr[0])
						product.SubRanking = "#" + utils.TrimAll(rank_arr[1])
					} else if len(rank_arr) == 1 {
						product.MainRanking = utils.TrimAll(rank_arr[0])
					}
				}
			}
			product.Size = strings.Join(sizes, ",")
		}

		var ProductValues amazon.ProductValues
		ProductValues.Price, _ = utils.ExtractNumberFromString(product.Price)
		ProductValues.Rating, _ = utils.ExtractNumberFromString(product.Rating)
		ProductValues.RatingsCount, _ = utils.ExtractNumberFromString(product.RatingsCount)
		ProductValues.ReviewCount, _ = utils.ExtractNumberFromString(product.ReviewCount)
		ProductValues.MainRanking, _ = utils.ExtractNumberFromString(product.MainRanking)
		ProductValues.SubRanking, _ = utils.ExtractNumberFromString(product.SubRanking)
		ProductValues.Availability, _ = utils.ExtractNumberFromString(product.Availability)
		if len(product.OtherSellersSpan) > 0 {
			ProductValues.OtherSellerCount, _ = utils.ExtractNumberFromString(product.OtherSellersSpan[0])
		}

		product.ProductValues = ProductValues

		product.DeliveryInfo = amazon.MerchantInfo2DeliveryInfo(product.MerchantInfo)

	})

	cy.OnError(func(r *colly.Response, err error) {
		onError = err
	})

	cy.OnResponse(func(r *colly.Response) {
		body := string(r.Body)
		if strings.Contains(body, "Sorry, we just need to make sure you're not a robot. For best results, please make sure your browser is accepting cookies.") {
			onError = fmt.Errorf("robot")
			return
		}
	})

	// Before making a request
	cy.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping
	err := cy.Visit(productURL)
	if err != nil {
		return nil, err
	}
	if onError != nil {
		return nil, onError
	}
	return product, nil
}

func getAmzTable(e *colly.HTMLElement, goquerySelector string) map[string]string {
	var table = make(map[string]string)
	e.ForEach(goquerySelector, func(i int, e *colly.HTMLElement) {
		e.ForEach("tr", func(i int, e *colly.HTMLElement) {
			td := e.DOM.Find("td")
			// 过滤掉td中的style元素
			td.Find("style").Remove()
			td.Find("script").Remove()
			table[utils.TrimAll(e.ChildText("th"))] = utils.TrimAll(td.Text())
		})
	})
	return table
}
