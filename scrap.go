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


func fetchURL(url, ean string) ([]byte, error) {
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
		return nil, fmt.Errorf("product %v not found at %v", ean, url)
	default:
		return nil, fmt.Errorf("error while processing product %v at while processing %v, received code %v", ean, url, resp.StatusCode)
	}
}


// Fetcher query something (URL, database, ...) with EAN, and return the Product stored or scrapped
type Fetcher interface {
	Fetch(ean string) (*Product, error)
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

// Fetcher for upcitemdb.com
var UpcItemDbFetcher upcItemDbURL
// Fetcher for openfoodfacts.org (using json api)
var OpenFoodFactsFetcher openFoodFactsURL


// Fetchers is a list of default fetchers already implemented.
// Currently supported websited:
// - upcitemdb
// - openfoodfacts
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
	fetchers = []Fetcher{UpcItemDbFetcher, OpenFoodFactsFetcher}
}

func (f upcItemDbURL) Fetch(ean string) (*Product, error) {
	url := f.fullURL(ean)
	body, err := fetchURL(url, ean)
	if err != nil {
		return nil, err
	}
	p := f.parseBody(body)
	if p == nil {
		return nil, fmt.Errorf("Cound not extract data from html page at %v for Product %v", url, ean)
	}
	p.EAN = ean
	p.URL = url
	return p, nil

}

func (f upcItemDbURL) parseBody(b []byte) *Product {
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	p := Product{}
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
								return
							}
						}
					}
				case "img":
					for _, attr := range c.Attr {
						if attr.Key == "src" {
							p.ImageURL = attr.Val
							return
						}
					}
				default:
					fn(c)
				}
			}
		}
	}
	fn(doc)
	if p.Name == "" {
		return nil
	}
	return &p
}

func (f openFoodFactsURL) Fetch(ean string) (*Product, error) {
	url := f.fullURL(ean)

	body, err := fetchURL(url, ean)
	if err != nil {
		return nil, err
	}
	var v interface{}
	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}

	m := v.(map[string]interface{})
	if status := m["status"].(float64); status != 1. { // 1 == product found
		return nil, fmt.Errorf("product %v not found at %v", ean, url)
	}
	productIntf, ok := m["product"]
	if !ok {
		return nil, fmt.Errorf("no product field found in json from %v", url)
	}
	product, ok := productIntf.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no product map found in json from %v", url)
	}
	nameIntf, ok := product["product_name"]
	if !ok {
		return nil, fmt.Errorf("no product_name field found in json from %v", url)
	}
	name, ok := nameIntf.(string)
	if !ok {
		return nil, fmt.Errorf("product_name from %v is not a string", url)
	}
	imageURLIntf, ok := product["image_front_url"]
	var imageURL string
	if !ok {
		imageURL = ""
	} else {
		imageURL, ok = imageURLIntf.(string)
		if !ok {
			return nil, fmt.Errorf("image_front_url from %v is not a string", url)
		}
	}

	return &Product{URL: url, EAN: ean, Name: name, ImageURL: imageURL}, nil
}


// Scrap a Product data bases on its EAN with default Fetchers
// All Default Fetchers are executed in goroutines
// Return the Product if it is found on one site (the fastest).
func Scrap(ean string) (*Product, error) {
	if !eancheck.Valid(ean) {
		return &Product{}, fmt.Errorf("invalid EAN %v", ean)
	}
	type prodErr struct {
		p   *Product
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
		} else if pe.p != nil {
			return pe.p, nil
		}
	}
	return nil, fmt.Errorf("no product found because of the following errors: %v", len(errors))
}
