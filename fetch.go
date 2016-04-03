package recycleme

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	eancheck "github.com/nicholassm/go-ean"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Product struct {
	EAN         string // EAN number for the Product
	Name        string // Name of the Product
	URL         string // URL where the details of the Product were found
	ImageURL    string // URL where to find an image of the Product
	WebsiteURL  string // URL where to find the details of the Product
	WebsiteName string // Website name
}

func (p Product) String() string {
	s := fmt.Sprintf("%v (%v) at %v (%v)", p.Name, p.EAN, p.URL, p.WebsiteURL)
	if p.ImageURL != "" {
		s += fmt.Sprintf("\n\tImage: %v", p.ImageURL)
	}
	return s
}

func (p Product) JSON() ([]byte, error) {
	return json.Marshal(p)
}

type memoryBlacklistDB struct {
	blacklisted map[string]struct{}
	sync.Mutex
}

func (b *memoryBlacklistDB) Add(url string) {
	b.Lock()
	b.blacklisted[url] = struct{}{}
	b.Unlock()
}

func (b *memoryBlacklistDB) Contains(url string) bool {
	_, ok := b.blacklisted[url]
	return ok
}

type BlacklistDB interface {
	Contains(string) bool
	Add(string)
}

var Blacklist = &memoryBlacklistDB{blacklisted: make(map[string]struct{})}

// Fetcher query something (URL, database, ...) with EAN, and return the Product stored or scrapped
type Fetcher interface {
	Fetch(ean string) (Product, error)
	IsURLValidForEAN(url, ean string) bool
}

// FetchableURL is a base struct to fetch websites
// URL that can be used by fetchers, it must be a format string, the %s or %v will be replaced by the EAN
// WebsiteName is the corporate name given to the website to be fetched, for prettier printing
type FetchableURL struct {
	URL         string
	WebsiteName string
}

// Create a new FetchableURL, checking that it contains the correct format to place the EAN in the URL
func NewFetchableURL(url string, website string) (FetchableURL, error) {
	if !strings.Contains(url, "%s") && !strings.Contains(url, "%v") {
		return FetchableURL{}, fmt.Errorf("URL %v does not containt format string to insert EAN", url)
	}

	return FetchableURL{URL: url, WebsiteName: website}, nil
}

func fullURL(url, ean string) string {
	return fmt.Sprintf(url, ean)
}

var client = http.Client{
	Timeout: time.Duration(15 * time.Second),
}

func fetchURL(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		body, err := ioutil.ReadAll(resp.Body)
		return body, err
	case 404:
		return nil, errNotFound
	default:
		return nil, fmt.Errorf("error while processing product %v, received code %v", url, resp.StatusCode)
	}
}

type upcItemDbURL FetchableURL

// Fetcher for upcitemdb.com
var UpcItemDbFetcher = upcItemDbURL{URL: "http://www.upcitemdb.com/upc/%s", WebsiteName: "UPCItemDB"}

type openFoodFactsURL FetchableURL

// Fetcher for openfoodfacts.org (using json api)
var OpenFoodFactsFetcher = openFoodFactsURL{URL: "http://fr.openfoodfacts.org/api/v0/produit/%s.json", WebsiteName: "OpenFoodFacts"}

type isbnSearchURL FetchableURL

// Fetcher for isbnsearch.org (using json api)
var IsbnSearchFetcher = isbnSearchURL{URL: "http://www.isbnsearch.org/isbn/%s", WebsiteName: "ISBNSearch"}

type amazonURL struct {
	endPoint                           string
	WebsiteName                        string
	AccessKey, SecretKey, AssociateTag string
}

// Fetcher for ean-search.org (using json api)
var AmazonFetcher amazonURL

func newAmazonURLFetcher() (amazonURL, error) {
	fetcher := amazonURL{}
	var accessOk, secretOk, associateTagOk bool
	fetcher.AccessKey, accessOk = os.LookupEnv("RECYCLEME_ACCESS_KEY")
	fetcher.SecretKey, secretOk = os.LookupEnv("RECYCLEME_SECRET_KEY")
	fetcher.AssociateTag, associateTagOk = os.LookupEnv("RECYCLEME_ASSOCIATE_TAG")
	if accessOk && secretOk && associateTagOk {
		fetcher.endPoint = "webservices.amazon.fr"
		fetcher.WebsiteName = "Amazon.fr"
		return fetcher, nil
	}
	return fetcher, errors.New("Missing either RECYCLEME_ACCESS_KEY, RECYCLEME_SECRET_KEY or RECYCLEME_ASSOCIATE_TAG in environment. AmazonFetcher will not be used")
}

func (f upcItemDbURL) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
}

func (f upcItemDbURL) Fetch(ean string) (Product, error) {
	url := fullURL(f.URL, ean)
	if Blacklist.Contains(url) {
		return Product{}, NewProductError(ean, url, errBlacklisted)
	}
	body, err := fetchURL(url)
	if err != nil {
		return Product{}, NewProductError(ean, url, err)
	}
	p, err := f.parseBody(body)
	if err != nil {
		return p, NewProductError(ean, url, err)
	}
	p.EAN = ean
	p.URL = url
	p.WebsiteURL = url
	p.WebsiteName = f.WebsiteName
	return p, nil

}

func (f upcItemDbURL) parseBody(b []byte) (Product, error) {
	doc, err := html.Parse(bytes.NewReader(b))
	p := Product{}
	if err != nil {
		return p, err
	}
	var fn func(*html.Node)
	fn = func(n *html.Node) {
		// printText = printText || (n.Type == html.ElementNode && n.Data == "b")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Looking for <p class="detailtitle">....<b>$PRODUCT_NAME</b></p>
			if c.Type == html.ElementNode {
				switch c.Data {
				case "p":
					if len(c.Attr) == 1 {
						classAttr := c.Attr[0]
						if classAttr.Val == "detailtitle" {
							txt := c.FirstChild.NextSibling.FirstChild
							if txt.Type == html.TextNode {
								p.Name = txt.Data
							}
						}
					}
				case "img":
					isProduct := false
					for _, attr := range c.Attr {
						if attr.Key == "class" && strings.Contains(attr.Val, "product") {
							isProduct = true
						}
						if attr.Key == "src" && isProduct && len(p.ImageURL) == 0 {
							p.ImageURL = attr.Val
						}
					}
				default:
					if p.ImageURL != "" && p.Name != "" {
						return
					}
					fn(c)
				}
			}
		}
	}
	fn(doc)
	if p.Name == "" {
		return p, errNotFound
	}
	return p, nil
}

func (f openFoodFactsURL) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
}

func (f openFoodFactsURL) Fetch(ean string) (Product, error) {
	url := fullURL(f.URL, ean)
	if Blacklist.Contains(url) {
		return Product{}, NewProductError(ean, url, errBlacklisted)
	}
	p := Product{}
	body, err := fetchURL(url)
	if err != nil {
		return Product{}, NewProductError(ean, url, err)
	}
	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return p, NewProductError(ean, url, err)
	}

	m := v.(map[string]interface{})
	if status := m["status"].(float64); status != 1. {
		// 1 == product found
		return p, NewProductError(ean, url, errNotFound)
	}
	productIntf, ok := m["product"]
	if !ok {
		return p, NewProductError(ean, url, fmt.Errorf("no product field found in json"))
	}
	product, ok := productIntf.(map[string]interface{})
	if !ok {
		return p, NewProductError(ean, url, fmt.Errorf("no product map found in json"))
	}
	nameIntf, ok := product["product_name"]
	if !ok {
		return p, NewProductError(ean, url, fmt.Errorf("no product_name field found in json"))
	}
	name, ok := nameIntf.(string)
	if !ok {
		return p, NewProductError(ean, url, fmt.Errorf("product_name is not a string"))
	}
	imageURLIntf, ok := product["image_front_url"]
	var imageURL string
	if !ok {
		imageURL = ""
	} else {
		imageURL, ok = imageURLIntf.(string)
		if !ok {
			return p, NewProductError(ean, url, fmt.Errorf("image_front_url is not a string"))
		}
	}
	websiteURL := fmt.Sprintf("http://fr.openfoodfacts.org/produit/%s/", ean)
	return Product{URL: url, EAN: ean, Name: name, ImageURL: imageURL, WebsiteURL: websiteURL, WebsiteName: f.WebsiteName}, nil
}

func (f isbnSearchURL) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
}

func (f isbnSearchURL) Fetch(ean string) (Product, error) {
	url := fullURL(f.URL, ean)
	if Blacklist.Contains(url) {
		return Product{}, NewProductError(ean, url, errBlacklisted)
	}
	body, err := fetchURL(url)
	if err != nil {
		return Product{}, NewProductError(ean, url, err)
	}
	p, err := f.parseBody(body)
	if err != nil {
		return p, NewProductError(ean, url, fmt.Errorf("could not extract data from html page"))
	}
	p.EAN = ean
	p.URL = url
	p.WebsiteURL = url
	p.WebsiteName = f.WebsiteName
	return p, nil
}

func (f isbnSearchURL) parseBody(b []byte) (Product, error) {
	doc, err := html.Parse(bytes.NewReader(b))
	p := Product{}
	if err != nil {
		return p, err
	}
	var fn func(*html.Node)
	fn = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Looking for <div class="bookinfo"><h2>$PRODUCT_NAME</h2></div>
			if c.Type == html.ElementNode {
				switch c.Data {
				case "div":
					if p.Name != "" {
						return
					}
					if len(c.Attr) == 1 {
						classAttr := c.Attr[0]
						if classAttr.Val == "bookinfo" {
							txt := c.FirstChild.NextSibling.FirstChild
							if txt.Type == html.TextNode {
								p.Name = txt.Data
								return
							}
						}
					}
				case "img":
					if p.ImageURL != "" {
						return
					}
					for _, attr := range c.Attr {
						if attr.Key == "src" {
							p.ImageURL = attr.Val
							return
						}
					}
				}
				fn(c)
			}
		}
	}
	fn(doc)
	if p.Name == "" {
		return p, errNotFound
	}
	return p, nil
}

// amazonItemSearchResponse is the base xml response, only keep needed fields (maybe more will be added later) in a "flat" struct.
type amazonItemSearchResponse struct {
	TotalResults uint          `xml:"Items>TotalResults"`
	Item         []amazonItem  `xml:"Items>Item"`
	IsValid      string        `xml:"Items>Request>IsValid"`
	RequestID    string        `xml:"OperationRequest>RequestId"`
	Errors       []amazonError `xml:"Items>Request>Errors>Error"`
}

type amazonError struct {
	Code    string
	Message string
}

func (e *amazonError) Error() string {
	return fmt.Sprintf("error from amazon: Code: %s, Message: %s", e.Code, e.Message)
}

type amazonItem struct {
	Title          string `xml:"ItemAttributes>Title"`
	ASIN           string
	DetailPageURL  string
	SmallImageURL  string `xml:"SmallImage>URL"`
	MediumImageURL string `xml:"MediumImage>URL"`
	LargeImageURL  string `xml:"LargeImage>URL"`
}

func (f amazonURL) IsURLValidForEAN(url, ean string) bool {
	return f.endPoint == url
}

func (f amazonURL) buildURL(ean string) (string, error) {
	params := url.Values{}
	params.Set("AWSAccessKeyId", f.AccessKey)
	params.Set("AssociateTag", f.AssociateTag)
	params.Set("Service", "AWSECommerceService")
	params.Set("Operation", "ItemSearch")
	params.Set("Timestamp", time.Now().Format(time.RFC3339))
	params.Set("SearchIndex", "All")
	params.Set("ResponseGroup", "Images,Small")
	params.Set("Keywords", ean)
	uri := "/onca/xml"
	toSign := fmt.Sprintf("GET\n%s\n%s\n%s", f.endPoint, uri, strings.Replace(params.Encode(), "+", "%20", -1))

	hasher := hmac.New(sha256.New, []byte(f.SecretKey))
	_, err := hasher.Write([]byte(toSign))
	if err != nil {
		return "", err
	}

	signed := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	params.Set("Signature", signed)

	url := fmt.Sprintf("http://%s%s?%s", f.endPoint, uri, params.Encode())
	return url, nil
}

func (f amazonURL) Fetch(ean string) (Product, error) {
	url, err := f.buildURL(ean)
	endPoint := fmt.Sprintf("%s/%s", f.endPoint, ean)
	if err != nil {
		return Product{}, NewProductError(ean, endPoint, err)
	}
	if Blacklist.Contains(url) {
		return Product{}, NewProductError(ean, endPoint, errBlacklisted)
	}
	body, err := fetchURL(url)
	if err != nil {
		return Product{}, NewProductError(ean, endPoint, err)
	}

	var response amazonItemSearchResponse
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return Product{}, NewProductError(ean, endPoint, err)
	}
	if response.IsValid == "False" {
		if len(response.Errors) == 0 {
			return Product{}, NewProductError(ean, endPoint, fmt.Errorf("invalid response for RequestId"+response.RequestID))
		}
		var errors []string
		for _, e := range response.Errors {
			errors = append(errors, e.Error())
		}
		return Product{}, NewProductError(ean, endPoint, fmt.Errorf(strings.Join(errors, "; ")))
	}

	if response.TotalResults == 0 {
		return Product{}, NewProductError(ean, endPoint, errNotFound)
	}
	if response.TotalResults > 1 {
		return Product{}, NewProductError(ean, endPoint, errTooManyProducts)
	}

	firstItem := response.Item[0]
	return Product{EAN: ean, URL: endPoint, Name: firstItem.Title, ImageURL: firstItem.LargeImageURL, WebsiteURL: firstItem.DetailPageURL, WebsiteName: f.WebsiteName}, nil
}

type DefaultFetcher struct {
	fetchers []Fetcher
}

// NewDefaultFetcher fetches data from a list of default fetchers already implemented.
// Currently supported websited:
// - upcitemdb
// - openfoodfacts
// - isbnsearch
// - amazon (if credentials are provided)
// TODO: should return a warning, or info, not an error.
func NewDefaultFetcher() (DefaultFetcher, error) {
	fetchers := []Fetcher{UpcItemDbFetcher, OpenFoodFactsFetcher, IsbnSearchFetcher}
	amazonFetcher, err := newAmazonURLFetcher()
	if err != nil {
		return DefaultFetcher{fetchers: fetchers}, err
	}
	AmazonFetcher = amazonFetcher
	fetchers = append(fetchers, amazonFetcher)
	return DefaultFetcher{fetchers: fetchers}, nil
}

func (f DefaultFetcher) IsURLValidForEAN(url, ean string) bool {
	for _, fetcher := range f.fetchers {
		if fetcher.IsURLValidForEAN(url, ean) {
			return true
		}
	}
	return false
}

// Fetch a Product data bases on its EAN with default Fetchers
// All Default Fetchers are executed in goroutines
// Return the Product if it is found on one site (the fastest).
func (f DefaultFetcher) Fetch(ean string) (Product, error) {
	if !eancheck.Valid(ean) {
		return Product{}, fmt.Errorf("invalid EAN %v", ean)
	}
	type prodErr struct {
		p   Product
		err error
	}

	c := make(chan prodErr)
	q := make(chan struct{})
	for _, f := range f.fetchers {
		go func(f Fetcher) {
			product, err := f.Fetch(ean)
			select {
			case <-q:
				return
			case c <- prodErr{product, err}:
				return
			}
		}(f)
	}

	defer close(q)
	errors := make([]error, 0, len(f.fetchers))
	i := 0
	for pe := range c {
		i++
		if pe.err != nil {
			errors = append(errors, pe.err)
			if i == len(f.fetchers) {
				break
			}
		} else {
			return pe.p, nil
		}
	}

	errStr := make([]string, 1, len(errors)+1)
	errStr[0] = ""
	for _, err := range errors {
		errStr = append(errStr, err.Error())
	}
	return Product{}, fmt.Errorf("no product found because of the following errors:%v", strings.Join(errStr, "\n - "))
}
