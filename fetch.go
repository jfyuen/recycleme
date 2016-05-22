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
	"golang.org/x/text/encoding/charmap"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Product struct {
	EAN         string `json:"ean" bson:"ean"`                   // EAN number for the Product
	Name        string `json:"name" bson:"name"`                 // Name of the Product
	URL         string `json:"url" bson:"url"`                   // URL where the details of the Product were found
	ImageURL    string `json:"image_url" bson:"image_url"`       // URL where to find an image of the Product
	WebsiteURL  string `json:"website_url" bson:"website_url"`   // URL where to find the details of the Product
	WebsiteName string `json:"website_name" bson:"website_name"` // Website name
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

// UpcItemDbFetcher for upcitemdb.com
var UpcItemDbFetcher, _ = NewFetchableURL("http://www.upcitemdb.com/upc/%s", "UPCItemDB", upcItemDbParser{})

// OpenFoodFactsFetcher for openfoodfacts.org (using json api)
var OpenFoodFactsFetcher, _ = NewFetchableURL("http://fr.openfoodfacts.org/api/v0/produit/%s.json", "OpenFoodFacts", openFoodFactsParser{})

// IsbnSearchFetcher for isbnsearch.org (using json api)
var IsbnSearchFetcher, _ = NewFetchableURL("http://www.isbnsearch.org/isbn/%s", "ISBNSearch", isbnSearchParser{})

// IGalerieFetcher for some unknown website: http://90.80.54.225/?img=161277&images=1859
var IGalerieFetcher, _ = NewFetchableURL("http://90.80.54.225/?search=%s", "90.80.54.225", iGalerieParser{baseURL: "http://90.80.54.225/"})

// StarrymartFetcher for starrymart.co.uk
var StarrymartFetcher, _ = NewFetchableURL("https://starrymart.co.uk/catalogsearch/result/?q=%s", "StarryMart", starrymartParser{})

// MisterPharmaWebFetcher for misterpharmaweb.com
var MisterPharmaWebFetcher, _ = NewFetchableURL("http://www.misterpharmaweb.com/recherche-resultats.php?search_in_description=1&ac_keywords=%s", "MisterPharmaWeb", misterPharmaWebParser{baseURL: "http://www.misterpharmaweb.com/"})

// MedisparFetcher for meddispar.fr
var MedisparFetcher, _ = NewFetchableURL("http://www.meddispar.fr/content/search?search_by_name=&search_by_cip=%s", "Medispar", medisparParser{baseURL: "http://www.meddispar.fr"})

// DigitEyesFetcher for digit-eyes.com
var DigitEyesFetcher, _ = NewFetchableURL("http://www.digit-eyes.com/upcCode/%s.html", "Digit-Eyes", digitEyesParser{})

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

type HTMLParser interface {
	ParseBody(b []byte) (Product, error)
}

// FetchableURL is a base struct to fetch websites
// URL that can be used by fetchers, it must be a format string, the %s or %v will be replaced by the EAN
// WebsiteName is the corporate name given to the website to be fetched, for prettier printing
type FetchableURL struct {
	URL         string
	WebsiteName string
	HTMLParser
}

// Create a new FetchableURL, checking that it contains the correct format to place the EAN in the URL
func NewFetchableURL(url string, website string, parser HTMLParser) (FetchableURL, error) {
	if !strings.Contains(url, "%s") && !strings.Contains(url, "%v") {
		return FetchableURL{}, fmt.Errorf("URL %v does not containt format string to insert EAN", url)
	}

	return FetchableURL{URL: url, WebsiteName: website, HTMLParser: parser}, nil
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

func (f FetchableURL) Fetch(ean string, db BlacklistDB) (Product, error) {
	url := fullURL(f.URL, ean)
	return withCheckInBlacklist(db, ean, url, func() (Product, error) {
		body, err := fetchURL(url)
		if err != nil {
			return Product{}, err
		}
		p, err := f.ParseBody(body)
		if err != nil {
			return p, err
		}
		p.EAN = ean
		p.URL = url
		if p.WebsiteURL == "" {
			p.WebsiteURL = url
		}
		p.WebsiteName = f.WebsiteName
		return p, nil
	})
}

func (f FetchableURL) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
}

type iGalerieParser struct{ baseURL string }

func (f iGalerieParser) ParseBody(b []byte) (Product, error) {
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
							if c.FirstChild != nil && c.FirstChild.NextSibling != nil && c.FirstChild.NextSibling.FirstChild != nil {
								txt := c.FirstChild.NextSibling.FirstChild
								if txt.Type == html.TextNode && !strings.Contains(txt.Data, "1 image trouv") {
									err = errTooManyProducts
									return
								}
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
								imageURL := strings.Replace(attr.Val, "background:url(/getimg.php?img=", "", 1)
								imageURL = strings.Replace(imageURL, ") no-repeat center", "", 1)
								imageURL = strings.Split(f.baseURL, "?")[0] + "albums/" + imageURL
								p.ImageURL = imageURL
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

type upcItemDbParser struct{}

func (f upcItemDbParser) ParseBody(b []byte) (Product, error) {
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
						if classAttr.Val == "detailtitle" && c.FirstChild != nil && c.FirstChild.NextSibling != nil && c.FirstChild.NextSibling.FirstChild != nil {
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

type openFoodFactsParser struct{}
type openFoodFactsJSON struct {
	EAN     string `json:"code"`
	Product struct {
		Name     string `json:"product_name"`
		ImageURL string `json:"image_front_url"`
	}
	Status int `json:"status"`
}

func (f openFoodFactsParser) ParseBody(body []byte) (Product, error) {
	var v openFoodFactsJSON
	p := Product{}
	err := json.Unmarshal(body, &v)
	if err != nil {
		return p, err
	}
	if v.Status != 1 {
		return p, errNotFound
	}
	p.Name = v.Product.Name
	p.ImageURL = v.Product.ImageURL
	p.WebsiteURL = fmt.Sprintf("http://fr.openfoodfacts.org/produit/%s/", v.EAN)
	return p, nil
}

type isbnSearchParser struct{}

func (f isbnSearchParser) ParseBody(b []byte) (Product, error) {
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

type amazonURL struct {
	endPoint                           string
	WebsiteName                        string
	AccessKey, SecretKey, AssociateTag string
}

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

func (f amazonURL) IsURLValidForEAN(url, ean string) bool {
	return f.endPoint+"/"+ean == url
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
	p, err := withCheckInBlacklist(db, ean, endPoint, func() (Product, error) {
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

type mgoLocalProductDB struct {
	mgoDB
	colName string
}

func NewMgoLocalProductDB(s *mgo.Session, colPrefix string) *mgoLocalProductDB {
	return &mgoLocalProductDB{mgoDB: mgoDB{session: s}, colName: colPrefix + "local_products"}
}

func (db mgoLocalProductDB) Fetch(ean string, blacklist BlacklistDB) (Product, error) {
	var foundProduct Product
	err := withMgoSession(db.session, func(s *mgo.Session) error {
		var products []Product
		if err := s.DB("").C(db.colName).Find(bson.M{"ean": ean}).All(&products); err != nil {
			return newProductError(ean, "/local/", err)
		}
		if len(products) == 0 {
			return newProductError(ean, "/local/", errNotFound)
		}
		var err error
		for _, p := range products {
			url := "/local/" + p.WebsiteName + "/" + p.EAN
			product, err := withCheckInBlacklist(blacklist, ean, url, func() (Product, error) {
				return p, nil
			})
			if err != nil {
				continue
			}
			p.URL = url
			foundProduct = product
			return nil
		}
		return err
	})
	return foundProduct, err
}

func (db mgoLocalProductDB) IsURLValidForEAN(myURL, ean string) bool {
	if !strings.HasPrefix(myURL, "/local/") {
		return false
	}
	isValid := false
	// Looking for WebsiteName in /local/WebsiteName/EAN
	websiteName := strings.Replace(strings.Replace(strings.Replace(myURL, "local", "", -1), ean, "", -1), "/", "", -1)
	err := withMgoSession(db.session, func(s *mgo.Session) error {
		count, err := s.DB("").C(db.colName).Find(bson.M{"ean": ean, "website_name": websiteName}).Count()
		if err != nil {
			return err
		}
		fmt.Println(count)
		if count == 1 {
			isValid = true
		}
		return nil
	})
	if err != nil {
		return false
	}
	return isValid
}

type starrymartParser struct{}

func (f starrymartParser) ParseBody(b []byte) (Product, error) {
	doc, err := html.Parse(bytes.NewReader(b))
	p := Product{}
	if err != nil {
		return p, err
	}
	var fn func(*html.Node)
	fn = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				if c.Data == "div" {
					if len(c.Attr) == 1 && c.Attr[0].Key == "class" && c.Attr[0].Val == "item-img-info" {
						if c.FirstChild != nil && c.FirstChild.NextSibling != nil {
							a := c.FirstChild.NextSibling
							for i := range a.Attr {
								params := a.Attr[i]
								switch params.Key {
								case "href":
									p.WebsiteURL = params.Val
								case "title":
									p.Name = params.Val
								}

							}
							if a.FirstChild != nil && a.FirstChild.NextSibling != nil && a.FirstChild.NextSibling.FirstChild != nil && a.FirstChild.NextSibling.FirstChild.NextSibling != nil {
								img := a.FirstChild.NextSibling.FirstChild.NextSibling
								for i := range img.Attr {
									params := img.Attr[i]
									switch params.Key {
									case "src":
										p.ImageURL = params.Val

									}
								}
							}
						}
						return
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

type misterPharmaWebParser struct {
	baseURL string
}

func (f misterPharmaWebParser) ParseBody(b []byte) (Product, error) {
	charsetReader := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
	doc, err := html.Parse(charsetReader) // charset is ISO-8859-1
	p := Product{}
	if err != nil {
		return p, err
	}
	var fn func(*html.Node)
	fn = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				if c.Data == "img" {
					isProduct := false
					for _, attr := range c.Attr {
						if attr.Key == "class" && attr.Val == "lazy" {
							isProduct = true
						} else if !isProduct {
							break
						}
						if attr.Key == "data-src" && isProduct && len(p.ImageURL) == 0 {
							p.ImageURL = f.baseURL + attr.Val
						}
						if attr.Key == "alt" && isProduct && len(p.Name) == 0 {
							p.Name = attr.Val
						}
					}
					if p.Name != "" && isProduct {
						if len(c.Parent.Attr) == 1 && c.Parent.Attr[0].Key == "href" {
							p.WebsiteURL = c.Parent.Attr[0].Val
						}
						return
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

type medisparParser struct {
	baseURL string
}

func (f medisparParser) ParseBody(b []byte) (Product, error) {
	charsetReader := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
	doc, err := html.Parse(charsetReader) // charset is ISO-8859-1
	p := Product{}
	if err != nil {
		return p, err
	}
	var fn func(*html.Node)
	fn = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				if c.Data == "a" {
					isProduct := false
					name := ""
					url := ""
					for _, attr := range c.Attr {
						if attr.Key == "class" && attr.Val == "drug_title" {
							isProduct = true
						}
						if attr.Key == "title" {
							name = strings.Replace(attr.Val, "  ", " ", -1)
							name = strings.Replace(name, "\n", "", -1)
						}
						if attr.Key == "href" {
							url = attr.Val
						}
					}
					if isProduct && name != "" && url != "" {
						p.Name = name
						if !strings.Contains(url, "http") {
							url = f.baseURL + url
						}
						p.WebsiteURL = url
						return
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

type digitEyesParser struct{}

func (f digitEyesParser) ParseBody(b []byte) (Product, error) {
	charsetReader := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
	doc, err := html.Parse(charsetReader) // charset is ISO-8859-1
	p := Product{}
	if err != nil {
		return p, err
	}
	var fn func(*html.Node)
	fn = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				if c.Data == "img" {
					imageURL := ""
					for _, attr := range c.Attr {
						if attr.Key == "src" {
							imageURL = attr.Val
						}
						if attr.Key == "alt" && strings.Contains(attr.Val, "image of ") {
							p.Name = strings.Replace(attr.Val, "image of ", "", -1)
							p.ImageURL = imageURL
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

type DefaultFetcher struct {
	fetchers []Fetcher
}

// NewDefaultFetcher fetches data from a list of default fetchers already implemented.
// Currently supported websites:
// - upcitemdb
// - openfoodfacts
// - isbnsearch
// - iGalerie (some random IP on internet)
// - amazon (if credentials are provided)
// - StarryMart
// - MisterPharmaWeb
// - Meddispar
// - Digit-Eyes
// - more fetchers are provided as arguments (local database, ...)
// TODO: should return a warning, or info, not an error.
func NewDefaultFetcher(otherFetchers ...Fetcher) (DefaultFetcher, error) {
	fetchers := []Fetcher{UpcItemDbFetcher, OpenFoodFactsFetcher, IsbnSearchFetcher, IGalerieFetcher, StarrymartFetcher, MisterPharmaWebFetcher, MedisparFetcher, DigitEyesFetcher}
	for _, f := range otherFetchers {
		fetchers = append(fetchers, f)
	}
	amazonFetcher, err := newAmazonURLFetcher()
	if err != nil {
		return DefaultFetcher{fetchers: fetchers}, err
	}
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
