package scrape

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gin/APIClient"
	"gin/amazon"
	"gin/utils"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	fhttp "github.com/saucesteals/fhttp"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"

func ExtractBrandName(text string) string {
	pattern := regexp.MustCompile(`Visit the (.*?) Store`)
	match := pattern.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1]
	}
	return text
}

type recs_list struct {
	ID          string `json:"id"`
	MetadataMap struct {
		RenderZgRank                           string `json:"render.zg.rank"`
		RenderZgBsmsCurrentSalesRank           string `json:"render.zg.bsms.currentSalesRank"`
		RenderZgBsmsPercentageChange           string `json:"render.zg.bsms.percentageChange"`
		RenderZgBsmsTwentyFourHourOldSalesRank string `json:"render.zg.bsms.twentyFourHourOldSalesRank"`
		DisablePercolateLinkParams             string `json:"disablePercolateLinkParams"`
	} `json:"metadataMap"`
	LinkParameters struct {
	} `json:"linkParameters"`
}

func GetAmzProductListV2(_url, proxy string) ([]amazon.CategoryRank, error) {
	useragent := browser.Computer()
	client, _ := APIClient.New(proxy)
	client.UserAgent = useragent

	var asins []amazon.CategoryRank
	var nextPage string
	visit := func(visitUrl string) error {
		var base *url.URL
		base, err := url.Parse(visitUrl)
		if err != nil {
			return err
		}
		AbsoluteURL := func(path string) string {
			u, err := url.Parse(path)
			if err != nil {
				return ""
			}
			return base.ResolveReference(u).String()
		}
		fmt.Printf("开始爬取： %s\n", visitUrl)
		rsp, err := HTTPGet(client, visitUrl)
		if err != nil {
			return err
		}
		body := string(rsp)
		if strings.Contains(body, "Sorry, we just need to make sure you're not a robot. For best results, please make sure your browser is accepting cookies.") {
			return fmt.Errorf("robot")
		}
		// html 标签解析
		rd := bytes.NewReader(rsp)
		doc, err := goquery.NewDocumentFromReader(rd)
		if err != nil {
			return err
		}
		var onError error
		OnHTML(doc, "div.p13n-desktop-grid", func(e *goquery.Selection) {
			attr, exists := e.Attr("data-client-recs-list")
			if exists {
				var recs_lists []recs_list
				err := json.Unmarshal([]byte(attr), &recs_lists)
				if err != nil {
					onError = err
					return
				}
				for _, rl := range recs_lists {
					rank := amazon.CategoryRank{Rank: rl.MetadataMap.RenderZgRank, ID: rl.ID}
					IdDiv := e.Find(fmt.Sprintf("#%s", rl.ID)).First()
					if src, b := IdDiv.Find("img").First().Attr("src"); b {
						rank.Img = AbsoluteURL(src)
					}
					if href, b := IdDiv.Find("a").First().Attr("href"); b {
						rank.Url = AbsoluteURL(href)
					}
					rank.Title = utils.TrimAll(IdDiv.Find("a>span>div").First().Text())
					rank.Price = utils.TrimAll(IdDiv.Find("span.a-color-price").First().Text())
					rank.Rating = utils.TrimAll(IdDiv.Find("div>a i.a-icon-star-small").First().Text())
					rank.RatingsCount = utils.TrimAll(IdDiv.Find("div>a span.a-size-small").First().Text())
					asins = append(asins, rank)
				}

			}
		})
		href := doc.Find("div.a-text-center ul li.a-normal a").First().AttrOr("href", "")
		if len(href) > 0 {
			nextPage = AbsoluteURL(href)
		}
		return onError

	}
	err := visit(_url)
	if err != nil {
		return nil, err
	}
	if len(nextPage) > 0 {
		err = visit(nextPage)
		if err != nil {
			return nil, err
		}
	}
	return asins, nil
}

func GetAmzProductList(_url, proxy string) ([]amazon.CategoryRank, error) {
	host := strings.Split(_url, "/")[2]
	cy := colly.NewCollector(
		colly.UserAgent(UserAgent),
		colly.AllowedDomains(host),
	)
	cy.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		RandomDelay: 5 * time.Second,
	})

	if len(proxy) > 0 { // 设置代理IP
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		cy.SetProxyFunc(func(_ *http.Request) (*url.URL, error) {
			return proxyURL, nil
		})
	}
	var onError error
	var asins []amazon.CategoryRank
	cy.OnHTML("div.p13n-desktop-grid", func(e *colly.HTMLElement) {
		attr, exists := e.DOM.Attr("data-client-recs-list")
		if exists {
			var recs_lists []recs_list
			err := json.Unmarshal([]byte(attr), &recs_lists)
			if err != nil {
				onError = err
				return
			}
			for _, rl := range recs_lists {
				rank := amazon.CategoryRank{Rank: rl.MetadataMap.RenderZgRank, ID: rl.ID}
				IdDiv := e.DOM.Find(fmt.Sprintf("#%s", rl.ID)).First()
				if src, b := IdDiv.Find("img").First().Attr("src"); b {
					rank.Img = e.Request.AbsoluteURL(src)
				}
				if href, b := IdDiv.Find("a").First().Attr("href"); b {
					rank.Url = e.Request.AbsoluteURL(href)
				}
				rank.Title = utils.TrimAll(IdDiv.Find("a>span>div").First().Text())
				rank.Price = utils.TrimAll(IdDiv.Find("span.a-color-price").First().Text())
				rank.Rating = utils.TrimAll(IdDiv.Find("div>a i.a-icon-star-small").First().Text())
				rank.RatingsCount = utils.TrimAll(IdDiv.Find("div>a span.a-size-small").First().Text())
				asins = append(asins, rank)
			}

		}
	})

	var visitCount = 1
	cy.OnHTML("div.a-text-center ul li.a-normal a", func(element *colly.HTMLElement) {
		src := element.Attr("href")
		if len(src) > 0 && visitCount > 0 {
			visitCount--
			element.Request.Visit(src)
		}
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
	err := cy.Visit(_url)
	if err != nil {
		return nil, err
	}
	if onError != nil {
		return nil, onError
	}
	return asins, nil
}
func OnHTML(e *goquery.Document, selector string, callback func(element *goquery.Selection)) {
	first := e.Find(selector).First()
	if first.Length() > 0 {
		callback(first)
	}
}
func ForEach(doc *goquery.Selection, selector string, callback func(i int, element *goquery.Selection)) {
	doc.Find(selector).Each(callback)
}
func getAmzTableV2(e *goquery.Selection, goquerySelector string) map[string]string {
	var table = make(map[string]string)
	ForEach(e, goquerySelector, func(i int, e *goquery.Selection) {
		ForEach(e, "tr", func(i int, e *goquery.Selection) {
			td := e.Find("td")
			// 过滤掉td中的style元素
			td.Find("style").Remove()
			td.Find("script").Remove()
			table[utils.TrimAll(e.Find("th").Text())] = utils.TrimAll(td.Text())
		})
	})
	return table
}

func HTTPGet(client *APIClient.ChromiumClient, _url string) ([]byte, error) {
	//nowTime := time.Now()
	//defer func() {
	//	fmt.Printf("[%d]ms [GET]: %s\n", time.Now().Sub(nowTime).Milliseconds(), _url)
	//}()
	req, err := fhttp.NewRequest("GET", _url, nil)
	resp, err := client.ClientDo(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	encoding := resp.Header["Content-Encoding"]
	content := resp.Header["Content-Type"]

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	Body := cycletls.DecompressBody(bodyBytes, encoding, content)
	return []byte(Body), nil
}
func GetAmzProductEx(host, asin, proxy string) (*amazon.Product, error) {
	productURL := fmt.Sprintf("https://%s/dp/%s?th=1&psc=1", host, asin)
	useragent := browser.Computer()
	client, _ := APIClient.New(proxy)
	client.UserAgent = useragent
	fmt.Printf("开始爬取： %s\n", productURL)
	rsp, err := HTTPGet(client, productURL)
	if err != nil {
		return nil, err
	}
	body := string(rsp)
	if strings.Contains(body, "Sorry, we just need to make sure you're not a robot. For best results, please make sure your browser is accepting cookies.") {
		return nil, fmt.Errorf("robot")
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("body is empty")
	}
	// html 标签解析
	rd := bytes.NewReader(rsp)
	doc, err := goquery.NewDocumentFromReader(rd)
	if err != nil {
		return nil, err
	}
	// 获取商品信息
	product := amazon.Product{}
	dbContainer := doc.Find("#dp-container").First()

	if dbContainer.Length() > 0 {
		product.ASIN = dbContainer.Find("div[data-asin]").AttrOr("data-asin", "")

		if len(product.ASIN) == 0 {
			return nil, fmt.Errorf("ASIN is empty")
		}

		product.Title = utils.TrimAll(dbContainer.Find("#productTitle").First().Text())
		product.Price = utils.TrimAll(dbContainer.Find("div.a-box-inner span.a-price span").First().Text())
		product.Brand = strings.TrimPrefix(dbContainer.Find("a#bylineInfo").First().Text(), "Brand: ")
		product.MerchantInfo = utils.TrimAll(dbContainer.Find("#merchant-info").First().Text())
		product.RatingsCount = utils.TrimAll(dbContainer.Find("#averageCustomerReviews_feature_div #acrCustomerReviewText").First().Text())
		product.Rating = utils.TrimAll(dbContainer.Find("#acrPopover").First().AttrOr("title", ""))
		product.ReviewCount = utils.TrimAll(dbContainer.Find("#askATFLink").First().Text())

		availability := dbContainer.Find("#availability span").First()
		product.Availability = utils.TrimAll(availability.Text())

		// 卖家数量
		var sellerSpan []string
		ForEach(dbContainer, "#olpLinkWidget_feature_div div.olp-text-box>span", func(i int, element *goquery.Selection) {
			v := utils.TrimAll(element.First().Text())
			if len(v) > 0 {
				sellerSpan = append(sellerSpan, v)
			}
		})
		product.OtherSellersSpan = sellerSpan
		table := getAmzTableV2(dbContainer, "table#productDetails_techSpec_section_1")
		if len(table) > 0 {
			product.Details = append(product.Details, table)
		}

		// 产品图片
		ForEach(dbContainer, "#leftCol li[data-csa-c-element-type=navigational] img", func(i int, element *goquery.Selection) {
			product.Images = append(product.Images, element.AttrOr("src", ""))
		})

		// 产品详情
		ForEach(dbContainer, "#tech table", func(i int, e *goquery.Selection) {
			var tb = make(map[string]string)
			ForEach(e, "tr", func(i int, e *goquery.Selection) {
				td := e.Find("td")
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
		ForEach(dbContainer, "#detailBullets_feature_div span.a-list-item", func(i int, e *goquery.Selection) {
			span := e.Find("span")
			if span.Length() >= 2 {
				tb[utils.TrimSpan(span.Eq(0).Text())] = utils.TrimSpan(span.Eq(1).Text())
			}
		})
		// 产品详情
		ForEach(dbContainer, "#detailBulletsWrapper_feature_div>ul", func(i int, e *goquery.Selection) {
			li := e.Find("li").First()
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

		table = getAmzTableV2(dbContainer, "table#productDetails_detailBullets_sections1")
		if len(table) > 0 {
			product.Details = append(product.Details, table)
		}
		product.MainRanking = dbContainer.Find("#SalesRank").Text()
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
				if asin, ok := detail["ASIN"]; ok {
					product.ASIN = asin
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

		if len(product.Brand) > 0 {
			product.Brand = ExtractBrandName(product.Brand)
		}

		if sellerName, ok := product.DeliveryInfo.Info["sellerName"]; ok {
			product.SellerNameContainsBrand = strings.Contains(strings.ToLower(product.Brand), strings.ToLower(sellerName))
			if !product.SellerNameContainsBrand {
				product.SellerNameContainsBrand = strings.Contains(strings.ToLower(sellerName), strings.ToLower(product.Brand))
			}
		}
	}

	return &product, nil
}

func GetAmzProduct(cy *colly.Collector, host, asin, proxy string) (*amazon.Product, error) {
	productURL := fmt.Sprintf("https://%s/dp/%s?th=1&psc=1", host, asin)
	useragent := browser.Computer()
	if cy == nil {
		cy = colly.NewCollector(
			colly.UserAgent(useragent),
			colly.AllowedDomains(host),
		)
		var proxyURL *url.URL
		var err error
		if len(proxy) > 0 {
			// 设置代理IP
			proxyURL, err = url.Parse(proxy)
			if err != nil {
				log.Fatal(err)
			}
			cy.SetProxyFunc(func(_ *http.Request) (*url.URL, error) {
				return proxyURL, nil
			})
		}

		// client.IMAPClient.Transport

		//cy.WithTransport(client.IMAPClient.Transport.(http.RoundTripper))
		/*		var Proxy func(*http.Request) (*url.URL, error)
				if proxyURL == nil {
					Proxy = http.ProxyFromEnvironment
				} else {
					Proxy = http.ProxyURL(proxyURL)
				}*/

		/*		tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					Proxy:           Proxy,
					DialContext: (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext,
					ForceAttemptHTTP2: true,
				}
				cy.WithTransport(tr)*/
		//extensions.RandomUserAgent(cy)
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
				if asin, ok := detail["ASIN"]; ok {
					product.ASIN = asin
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

		if len(product.Brand) > 0 {
			product.Brand = ExtractBrandName(product.Brand)
		}

		if sellerName, ok := product.DeliveryInfo.Info["sellerName"]; ok {
			product.SellerNameContainsBrand = strings.Contains(strings.ToLower(product.Brand), strings.ToLower(sellerName))
			if !product.SellerNameContainsBrand {
				product.SellerNameContainsBrand = strings.Contains(strings.ToLower(sellerName), strings.ToLower(product.Brand))
			}
		}

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
	if len(product.ASIN) == 0 {
		return nil, errors.New("len(product.ASIN)==0")
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

// MyFingerprinter 实现了 tlsfingerprint.Fingerprinter 接口，用于计算 TLS 指纹
type MyFingerprinter struct{}

func (f *MyFingerprinter) Fingerprint(transport *http.Transport) (string, error) {
	// 获取证书
	certs := transport.TLSClientConfig.Certificates

	// 计算所有证书的 SHA256 哈希值
	hash := sha256.New()
	for _, cert := range certs {
		hash.Write(cert.Certificate[0])
	}
	sum := hash.Sum(nil)

	// 返回十六进制格式的哈希值
	return hex.EncodeToString(sum), nil
}
