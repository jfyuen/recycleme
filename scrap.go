package recycleme

import (
	"bytes"
	"encoding/json"
	"fmt"
	eancheck "github.com/nicholassm/go-ean"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Product struct {
	EAN      string // EAN number for the Product
	Name     string // Name of the Product
	URL      string // URL where the details of the Product were found
	ImageURL string // URL where to find an image of the Product
}

func (p Product) String() string {
	s := fmt.Sprintf("%v (%v) at %v", p.Name, p.EAN, p.URL)
	if p.ImageURL != "" {
		s += fmt.Sprintf("\n\tImage: %v", p.ImageURL)
	}
	return s
}

func (p Product) Json() ([]byte, error) {
	return json.Marshal(p)
}

func fetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
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

// Fetcher query something (URL, database, ...) with EAN, and return the Product stored or scrapped
type Fetcher interface {
	Fetch(ean string) (Product, error)
}

// URL that can be fetched by fetchers, it must be a format string, the %s will be replaced by the EAN
type FetchableURL string

// Create a new FetchableURL, checking that it contains the correct format to place the EAN in the URL
func NewFetchableURL(url string) (FetchableURL, error) {
	if !strings.Contains(url, "%s") && !strings.Contains(url, "%v") {
		return FetchableURL(""), fmt.Errorf("URL %v does not containt format string to insert EAN", url)
	}

	return FetchableURL(url), nil
}

func (f FetchableURL) fullURL(ean string) string {
	return fmt.Sprintf(string(f), ean)
}

type upcItemDbURL struct {
	FetchableURL
}

type openFoodFactsURL struct {
	FetchableURL
}

type isbnSearchUrl struct {
	FetchableURL
}

// Fetcher for upcitemdb.com
var UpcItemDbFetcher upcItemDbURL

// Fetcher for openfoodfacts.org (using json api)
var OpenFoodFactsFetcher openFoodFactsURL

var IsbnSearchFetcher isbnSearchUrl

// Fetchers is a list of default fetchers already implemented.
// Currently supported websited:
// - upcitemdb
// - openfoodfacts
// - isbnsearch
var fetchers []Fetcher

func init() {
	fetchable, err := NewFetchableURL("http://www.upcitemdb.com/upc/%s")
	if err != nil {
		log.Fatal(err)
	}
	UpcItemDbFetcher = upcItemDbURL{fetchable}

	fetchable, err = NewFetchableURL("http://fr.openfoodfacts.org/api/v0/produit/%s.json")
	if err != nil {
		log.Fatal(err)
	}
	OpenFoodFactsFetcher = openFoodFactsURL{fetchable}

	fetchable, err = NewFetchableURL("http://www.isbnsearch.org/isbn/%s")
	if err != nil {
		log.Fatal(err)
	}
	IsbnSearchFetcher = isbnSearchUrl{fetchable}
	fetchers = []Fetcher{UpcItemDbFetcher, OpenFoodFactsFetcher, IsbnSearchFetcher}
}

func (f upcItemDbURL) Fetch(ean string) (Product, error) {
	url := f.fullURL(ean)
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
		//		printText = printText || (n.Type == html.ElementNode && n.Data == "b")
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
						if attr.Key == "class" && attr.Val == "product" {
							isProduct = true
						}
						if attr.Key == "src" && isProduct {
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

func (f openFoodFactsURL) Fetch(ean string) (Product, error) {
	url := f.fullURL(ean)
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
	if status := m["status"].(float64); status != 1. { // 1 == product found
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

	return Product{URL: url, EAN: ean, Name: name, ImageURL: imageURL}, nil
}

func (f isbnSearchUrl) Fetch(ean string) (Product, error) {
	url := f.fullURL(ean)
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
	return p, nil

}

func (f isbnSearchUrl) parseBody(b []byte) (Product, error) {
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

// Scrap a Product data bases on its EAN with default Fetchers
// All Default Fetchers are executed in goroutines
// Return the Product if it is found on one site (the fastest).
func Scrap(ean string) (Product, error) {
	if !eancheck.Valid(ean) {
		return Product{}, fmt.Errorf("invalid EAN %v", ean)
	}
	type prodErr struct {
		p   Product
		err error
	}

	c := make(chan prodErr)
	q := make(chan struct{})
	for _, f := range fetchers {
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
	errors := make([]error, 0, len(fetchers))
	i := 0
	for pe := range c {
		i += 1
		if pe.err != nil {
			errors = append(errors, pe.err)
			if i == len(fetchers) {
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
