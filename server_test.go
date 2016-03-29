package recycleme

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func init() {
	canSendMail = false
}

func (f FetchableURL) Fetch(ean string) (Product, error) {
	return Product{}, nil
}

func (f FetchableURL) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
}

func TestAddBlacklistHandler(t *testing.T) {
	nopFetcher := FetchableURL{URL: "http://www.example.com/%s/", WebsiteName: "Example.com"}
	data := url.Values{}
	data.Set("name", "test product")
	ean := "EAN_TEST"
	data.Set("ean", ean)
	url := fullURL(nopFetcher.URL, ean)
	data.Set("url", url)
	data.Set("website", nopFetcher.WebsiteName)

	req, err := http.NewRequest("POST", "/blacklist/add", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.New(ioutil.Discard, "", 0)
		Blacklist.AddBlacklistHandler(w, r, logger, nopFetcher)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "added"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	if !Blacklist.Contains(url) {
		t.Errorf("%v not added to blacklist", url)
	}
}
