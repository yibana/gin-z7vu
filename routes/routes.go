package routes

import (
	"encoding/json"
	"fmt"
	"gin/amazon"
	"gin/config"
	"gin/db"
	"gin/utils"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func Readme(c *gin.Context) {
	filePath := filepath.Join(".", "README.md")
	markdownContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}
	// 渲染模板，并将markdownContent变量替换为动态的Markdown内容
	c.HTML(200, "markdown.html", gin.H{
		"MarkdownContent": string(markdownContent),
		"title":           "README.md",
	})
}

func AllCategorys(c *gin.Context) {
	filepath := filepath.Join(".", "category.json")
	categorys, err := ioutil.ReadFile(filepath)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}
	c.Data(200, "application/json", categorys)
}
func GetProduct(c *gin.Context) {
	host := c.DefaultQuery("host", "www.amazon.ca")
	asin := c.DefaultQuery("asin", "B08MR2C1T7")

	productURL := fmt.Sprintf("https://%s/dp/%s?th=1&psc=1", host, asin)
	// Create a new collector
	cy := colly.NewCollector(
		colly.AllowedDomains(host),
	)

	if len(config.ProxyUrl) > 0 {
		// 设置代理IP
		proxyURL, err := url.Parse(config.ProxyUrl)
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
	cy.OnHTML("#dp-container", func(e *colly.HTMLElement) {
		//ioutil.WriteFile("product.html", []byte(string(e.Response.Body)), 0644)
		product.ASIN = e.ChildAttr("div[data-asin]", "data-asin")
		product.Title = e.ChildText("#productTitle")
		product.Price = e.ChildText("#corePriceDisplay_desktop_feature_div span.a-offscreen")
		product.Brand = strings.TrimPrefix(e.ChildText("a#bylineInfo"), "Brand: ")
		product.MerchantInfo = e.ChildText("#merchant-info")
		product.RatingsCount = e.ChildText("#averageCustomerReviews_feature_div #acrCustomerReviewText")
		product.Rating = e.ChildAttr("#acrPopover", "title")
		product.ReviewCount = e.ChildText("#askATFLink")

		table := GetTable(e, "table#productDetails_techSpec_section_1")
		if len(table) > 0 {
			product.Details = append(product.Details, table)
		}

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
		tb := make(map[string]string)
		e.ForEach("#detailBullets_feature_div span.a-list-item", func(i int, e *colly.HTMLElement) {
			span := e.DOM.Find("span")
			if span.Length() >= 2 {
				tb[utils.TrimSpan(span.Eq(0).Text())] = utils.TrimSpan(span.Eq(1).Text())
			}
		})
		if len(tb) > 0 {
			product.Details = append(product.Details, tb)
		}

		table = GetTable(e, "table#productDetails_detailBullets_sections1")
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
					if len(rank_arr) == 2 {
						product.MainRanking = rank_arr[0]
						product.SubRanking = "#" + rank_arr[1]
					} else if len(rank_arr) == 1 {
						product.MainRanking = rank_arr[0]
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
		product.ProductValues = ProductValues

	})

	// Before making a request
	cy.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping
	err := cy.Visit(productURL)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error Visit: %s", err.Error()))
		return
	}

	// Print the product details
	marshal, err := json.MarshalIndent(product, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error MarshalIndent: %s", err.Error()))
		return
	}
	// 将product保存到mongodb数据库
	db.MongoMangerInstance.SaveProduct(product)

	c.Data(200, "application/json", marshal)
}

func GetTable(e *colly.HTMLElement, goquerySelector string) map[string]string {
	var table = make(map[string]string)
	e.ForEach(goquerySelector, func(i int, e *colly.HTMLElement) {
		e.ForEach("tr", func(i int, e *colly.HTMLElement) {
			table[utils.TrimAll(e.ChildText("th"))] = utils.TrimAll(e.ChildText("td"))
		})
	})
	return table
}
