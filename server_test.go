package recycleme

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func (f FetchableURL) Fetch(ean string) (Product, error) {
	return Product{Name: "TEST", URL: fullURL(f.URL, ean), WebsiteName: f.WebsiteName, EAN: ean}, nil
}

func (f FetchableURL) IsURLValidForEAN(url, ean string) bool {
	return fullURL(f.URL, ean) == url
}

func createPostRequest(uri string, data url.Values) (*http.Request, error) {
	req, err := http.NewRequest("POST", uri, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

type mailTester struct {
	expectedSubject, expectedBody string
	err                           error
}

func (m *mailTester) sendMail(subject, body string) error {
	if subject != m.expectedSubject || body != m.expectedBody {
		m.err = fmt.Errorf("subject or body differ: got %v and %v, expected %v and %v", subject, body, m.expectedSubject, m.expectedBody)
	}
	return m.err
}

func TestAddBlacklistHandler(t *testing.T) {
	nopFetcher := FetchableURL{URL: "http://www.example.com/%s/", WebsiteName: "Example.com"}
	data := url.Values{}
	name := "test product"
	data.Set("name", name)
	ean := "EAN_TEST"
	data.Set("ean", ean)
	url := fullURL(nopFetcher.URL, ean)
	data.Set("url", url)
	data.Set("website", nopFetcher.WebsiteName)

	req, err := createPostRequest("/blacklist/add", data)
	if err != nil {
		t.Fatal(err)
	}

	m := &mailTester{expectedBody: fmt.Sprintf("Blacklisting %s.\n%s should be %s", url, ean, name), expectedSubject: ean + " blacklisted"}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.New(ioutil.Discard, "", 0)
		Blacklist.AddBlacklistHandler(w, r, logger, nopFetcher, m.sendMail)
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "added"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	if m.err != nil {
		t.Error(m.err)
	}

	if !Blacklist.Contains(url) {
		t.Errorf("%v not added to blacklist", url)
	}

	data.Set("url", "invalid")
	req, err = createPostRequest("/blacklist/add", data)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestThrowAwayHandler(t *testing.T) {
	ean := "4006381333634"
	req, err := http.NewRequest("GET", "/throwaway/"+ean, nil)
	if err != nil {
		t.Fatal(err)
	}

	nopFetcher := FetchableURL{URL: "http://www.example.com/%s/", WebsiteName: "Example.com"}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ThrowAwayHandler(w, r, nopFetcher)
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	url := fullURL(nopFetcher.URL, ean)
	expected := fmt.Sprintf(`{"Product":{"EAN":"%s","Name":"TEST","URL":"%s","ImageURL":"","WebsiteURL":"","WebsiteName":"%s"},"ThrowAway":{}}`, ean, url, nopFetcher.WebsiteName)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestMaterialsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/materials/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(MaterialsHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	var out map[string]Material
	err = json.Unmarshal(rr.Body.Bytes(), &out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(Materials) {
		t.Fatalf("got different size %v vs %v", len(out), len(Materials))
	}
	for k, v := range out {
		id, err := strconv.Atoi(k)
		if err != nil {
			t.Fatal(err)
		}
		m := Materials[id]
		if v.Name != m.Name {
			t.Errorf("got different values for key %v: %v vs %v", k, v.Name, m.Name)
		}
	}
}
