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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	eancheck "github.com/nicholassm/go-ean"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Product struct {
	EAN         string `json:"ean"`         // EAN number for the Product
	Name        string `json:"name"`        // Name of the Product
	URL         string `json:"url"`         // URL where the details of the Product were found
	ImageURL    string `json:"imageURL"`    // URL where to find an image of the Product
	WebsiteURL  string `json:"websiteURL"`  // URL where to find the details of the Product
	WebsiteName string `json:"websiteName"` // Website name
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

type BlacklistDB interface {
	Contains(url string) (bool, error)
	Add(url string) error
}
type mgoBlacklistDB struct {
	mgoDB
	blacklistColName string
}

func NewMgoBlacklistDB(s *mgo.Session, colPrefix string) *mgoBlacklistDB {
	return &mgoBlacklistDB{mgoDB: mgoDB{session: s}, blacklistColName: colPrefix + "blacklist"}
}

func (db mgoBlacklistDB) Contains(url string) (bool, error) {
	var r bool
	err := withMgoSession(db.session, func(s *mgo.Session) error {
		col := s.DB("").C(db.blacklistColName)
		n, err := col.Find(bson.M{"url": url}).Count()
		if err != nil {
			return err
		}
		r = n == 1
		return nil
	})
	return r, err
}

func (db mgoBlacklistDB) Add(url string) error {
	err := withMgoSession(db.session, func(s *mgo.Session) error {
		col := s.DB("").C(db.blacklistColName)
		if _, err := col.Upsert(bson.M{"url": url}, bson.M{"url": url}); err != nil {
			return err
		}
		return nil
	})
	return err
}

// Fetcher query something (URL, database, ...) with EAN, and return the Product stored or scrapped
// It should check if the requested URL is in the blacklist
type Fetcher interface {
	Fetch(ean string, db BlacklistDB) (Product, error)
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

// UpcItemDbFetcher for upcitemdb.com
var UpcItemDbFetcher = upcItemDbURL{URL: "http://www.upcitemdb.com/upc/%s", WebsiteName: "UPCItemDB"}

type openFoodFactsURL FetchableURL

// OpenFoodFactsFetcher for openfoodfacts.org (using json api)
var OpenFoodFactsFetcher = openFoodFactsURL{URL: "http://fr.openfoodfacts.org/api/v0/produit/%s.json", WebsiteName: "OpenFoodFacts"}

type isbnSearchURL FetchableURL

// IsbnSearchFetcher for isbnsearch.org (using json api)
var IsbnSearchFetcher = isbnSearchURL{URL: "http://www.isbnsearch.org/isbn/%s", WebsiteName: "ISBNSearch"}

type iGalerieFetcher FetchableURL

// IGalerieFetcher for some unknown website: http://90.80.54.225/?img=161277&images=1859
var IGalerieFetcher = iGalerieFetcher{URL: "http://90.80.54.225/?search=%s", WebsiteName: "90.80.54.225"}

type amazonURL struct {
	endPoint                           string
	WebsiteName                        string
	AccessKey, SecretKey, AssociateTag string
}

// AmazonFetcher for amazon associate (using xml api)
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

type innerFetchFunc func() (Product, error)

func withCheckInBlacklist(b BlacklistDB, ean, url string, fn innerFetchFunc) (Product, error) {
	if ok, err := b.Contains(url); err != nil {
		return Product{}, newProductError(ean, url, err)
	} else if ok {
		return Product{}, newProductError(ean, url, errBlacklisted)
	}

	p, err := fn()
	if err != nil {
		return Product{}, newProductError(ean, url, err)
	}
	return p, nil
}

func (f iGalerieFetcher) Fetch(ean string, db BlacklistDB) (Product, error) {
	url := fullURL(f.URL, ean)
	return withCheckInBlacklist(db, ean, url, func() (Product, error) {
		body, err := fetchURL(url)
		if err != nil {
			return Product{}, err
		}
		p, err := f.parseBody(body)
		if err != nil {
			return p, err
		}
		p.EAN = ean
		p.URL = url
		p.WebsiteURL = url
		p.WebsiteName = f.WebsiteName
		return p, nil
	})
}

func (f iGalerieFetcher) parseBody(b []byte) (Product, error) {
	doc, err := html.Parse(bytes.NewReader(b))
	p := Product{}
	if err != nil {
		return p, err
	}
	var fn func(*html.Node)
	fn = func(n *html.Node) {
		// printText = printText || (n.Type == html.ElementNode && n.Data == "b")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				switch c.Data {
				// Looking for <div id="search_result"><p id="search_result_img">1 image trouvée :</p></div>
				case "div":
					if len(c.Attr) == 1 {
						classAttr := c.Attr[0]
						if classAttr.Key == "id" && classAttr.Val == "search_result" {
							txt := c.FirstChild.NextSibling.FirstChild
							if txt.Type == html.TextNode && !strings.Contains(txt.Data, "1 image trouv") {
								err = errTooManyProducts
								return
							}
						}
					}
				// image link is stored in style attribute
				case "a":
					if len(c.Attr) == 3 {
						isProduct := false
						for _, attr := range c.Attr {
							if attr.Key == "class" && strings.Contains(attr.Val, "img_link") {
								isProduct = true
							}
							if attr.Key == "style" && isProduct && len(p.ImageURL) == 0 && strings.Contains(attr.Val, "background:url") {
								imageUrl := strings.Replace(attr.Val, "background:url(/getimg.php?img=", "", 1)
								imageUrl = strings.Replace(imageUrl, ") no-repeat center", "", 1)
								imageUrl = strings.Split(f.URL, "?")[0] + "albums/" + imageUrl
								p.ImageURL = imageUrl
							}
						}
					}
				}
				if p.ImageURL != "" {
					return
				}
				fn(c)
			}
		}
	}
	fn(doc)
	if err != nil {
		return p, err
	}
	if p.ImageURL == "" {
		return p, errNotFound
	}
	return p, nil
}

func (f iGalerieFetcher) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
}

func (f upcItemDbURL) Fetch(ean string, db BlacklistDB) (Product, error) {
	url := fullURL(f.URL, ean)
	return withCheckInBlacklist(db, ean, url, func() (Product, error) {
		body, err := fetchURL(url)
		if err != nil {
			return Product{}, err
		}
		p, err := f.parseBody(body)
		if err != nil {
			return p, err
		}
		p.EAN = ean
		p.URL = url
		p.WebsiteURL = url
		p.WebsiteName = f.WebsiteName
		return p, nil
	})
}

func (f upcItemDbURL) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
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

func (f openFoodFactsURL) Fetch(ean string, db BlacklistDB) (Product, error) {
	url := fullURL(f.URL, ean)
	return withCheckInBlacklist(db, ean, url, func() (Product, error) {
		p := Product{}
		body, err := fetchURL(url)
		if err != nil {
			return Product{}, err
		}
		var v interface{}
		err = json.Unmarshal(body, &v)
		if err != nil {
			return p, err
		}

		m := v.(map[string]interface{})
		if status := m["status"].(float64); status != 1. {
			// 1 == product found
			return p, errNotFound
		}
		productIntf, ok := m["product"]
		if !ok {
			return p, fmt.Errorf("no product field found in json")
		}
		product, ok := productIntf.(map[string]interface{})
		if !ok {
			return p, fmt.Errorf("no product map found in json")
		}
		nameIntf, ok := product["product_name"]
		if !ok {
			return p, fmt.Errorf("no product_name field found in json")
		}
		name, ok := nameIntf.(string)
		if !ok {
			return p, fmt.Errorf("product_name is not a string")
		}
		imageURLIntf, ok := product["image_front_url"]
		var imageURL string
		if !ok {
			imageURL = ""
		} else {
			imageURL, ok = imageURLIntf.(string)
			if !ok {
				return p, fmt.Errorf("image_front_url is not a string")
			}
		}
		websiteURL := fmt.Sprintf("http://fr.openfoodfacts.org/produit/%s/", ean)
		return Product{URL: url, EAN: ean, Name: name, ImageURL: imageURL, WebsiteURL: websiteURL, WebsiteName: f.WebsiteName}, nil
	})
}

func (f isbnSearchURL) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
}

func (f isbnSearchURL) Fetch(ean string, db BlacklistDB) (Product, error) {
	url := fullURL(f.URL, ean)
	return withCheckInBlacklist(db, ean, url, func() (Product, error) {
		body, err := fetchURL(url)
		if err != nil {
			return Product{}, err
		}
		p, err := f.parseBody(body)
		if err != nil {
			return p, fmt.Errorf("could not extract data from html page")
		}
		p.EAN = ean
		p.URL = url
		p.WebsiteURL = url
		p.WebsiteName = f.WebsiteName
		return p, nil
	})
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

func (f amazonURL) Fetch(ean string, db BlacklistDB) (Product, error) {
	url, err := f.buildURL(ean)
	endPoint := fmt.Sprintf("%s/%s", f.endPoint, ean)
	if err != nil {
		return Product{}, newProductError(ean, endPoint, err)
	}
	p, err := withCheckInBlacklist(db, ean, url, func() (Product, error) {
		body, err := fetchURL(url)
		if err != nil {
			return Product{}, err
		}

		var response amazonItemSearchResponse
		err = xml.Unmarshal(body, &response)
		if err != nil {
			return Product{}, err
		}
		if response.IsValid == "False" {
			if len(response.Errors) == 0 {
				return Product{}, fmt.Errorf("invalid response for RequestId" + response.RequestID)
			}
			var errors []string
			for _, e := range response.Errors {
				errors = append(errors, e.Error())
			}
			return Product{}, fmt.Errorf(strings.Join(errors, "; "))
		}

		if response.TotalResults == 0 {
			return Product{}, errNotFound
		}
		if response.TotalResults > 1 {
			return Product{}, errTooManyProducts
		}

		firstItem := response.Item[0]
		return Product{EAN: ean, URL: endPoint, Name: firstItem.Title, ImageURL: firstItem.LargeImageURL, WebsiteURL: firstItem.DetailPageURL, WebsiteName: f.WebsiteName}, nil
	})
	if err != nil {
		pErr := err.(*productError)
		pErr.URL = endPoint
	}
	return p, err
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
	fetchers := []Fetcher{UpcItemDbFetcher, OpenFoodFactsFetcher, IsbnSearchFetcher, IGalerieFetcher}
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
func (f DefaultFetcher) Fetch(ean string, db BlacklistDB) (Product, error) {
	if !eancheck.Valid(ean) {
		return Product{}, errInvalidEAN

	}
	type prodErr struct {
		p   Product
		err error
	}

	c := make(chan prodErr)
	q := make(chan struct{})
	for _, f := range f.fetchers {
		go func(f Fetcher) {
			product, err := f.Fetch(ean, db)
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
