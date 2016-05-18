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
	"sync"
	"testing"
)

type testFetcher FetchableURL

func (f testFetcher) Fetch(ean string, db BlacklistDB) (Product, error) {
	return Product{Name: "TEST", URL: fullURL(f.URL, ean), WebsiteName: f.WebsiteName, EAN: ean}, nil
}

func (f testFetcher) IsURLValidForEAN(url, ean string) bool {
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
	wg                            *sync.WaitGroup
}

func newMailTester(subject, body string) *mailTester {
	m := &mailTester{expectedSubject: subject, expectedBody: body, wg: &sync.WaitGroup{}}
	m.wg.Add(1)
	return m
}

func (m *mailTester) sendMail(subject, body string) error {
	if subject != m.expectedSubject || body != m.expectedBody {
		m.err = fmt.Errorf("subject or body differ: got %v and %v, expected %v and %v", subject, body, m.expectedSubject, m.expectedBody)
	}
	m.wg.Done()
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

	m := newMailTester(fmt.Sprintf(ean+" blacklisted"), fmt.Sprintf("Blacklisting %s.\n%s should be %s", url, ean, name))
	handler := AddBlacklistHandler{
		Logger:    log.New(ioutil.Discard, "", 0),
		Blacklist: blacklistDB,
		Fetcher:   nopFetcher,
		Mailer:    m.sendMail,
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	m.wg.Wait()
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

	if ok, err := blacklistDB.Contains(url); err != nil {
		t.Fatal(err)
	} else if !ok {
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

	nopFetcher := testFetcher{URL: "http://www.example.com/%s/", WebsiteName: "Example.com"}
	handler := ThrowAwayHandler{
		DB:          packageDB,
		BlacklistDB: blacklistDB,
		Fetcher:     nopFetcher,
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	url := fullURL(nopFetcher.URL, ean)
	expected := fmt.Sprintf(`{"product":{"ean":"%s","name":"TEST","url":"%s","image_url":"","website_url":"","website_name":"%s","materials":[]},"throwAway":{}}`, ean, url, nopFetcher.WebsiteName)
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
	handler := MaterialsHandler{
		DB: packageDB,
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	var out []Material
	err = json.Unmarshal(rr.Body.Bytes(), &out)
	if err != nil {
		t.Fatal(err)
	}
	materials, err := packageDB.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(materials) {
		t.Fatalf("got different size %v vs %v", len(out), len(materials))
	}
	for i, v := range out {
		expected := materials[i]
		if v.Name != expected.Name || v.ID != expected.ID {
			t.Errorf("got material %v, expected %v", v.Name, expected.Name)
		}
	}
}

func TestAddPackageHandler(t *testing.T) {
	data := url.Values{}
	ean := "5021991938818"
	data.Set("ean", ean)
	expectedMaterials := []Material{Material{ID: 1, Name: "Boîte carton"}, Material{ID: 2, Name: "Film plastique"}}
	materialsJSON, err := json.Marshal(expectedMaterials)
	if err != nil {
		t.Fatal(err)
	}
	data.Set("materials", string(materialsJSON))

	req, err := createPostRequest("/materials/add", data)
	if err != nil {
		t.Fatal(err)
	}
	m := newMailTester("Adding package for "+ean, fmt.Sprintf("Materials added to %v:\n%v", ean, "[{1 Boîte carton} {2 Film plastique}]"))
	handler := AddPackageHandler{
		Logger: log.New(ioutil.Discard, "", 0),
		DB:     packageDB,
		Mailer: m.sendMail,
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	m.wg.Wait()

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

	if v, err := packageDB.Get(ean); err != nil {
		if err == errPackageNotFound {
			t.Errorf("%v not added to packages", ean)
		} else {
			t.Fatal(err)
		}
	} else {
		if len(v.Materials) != len(expectedMaterials) {
			t.Errorf("expected %v packages, got %v", len(expectedMaterials), len(v.Materials))
		}

		for i := 0; i < 2; i++ {
			got := v.Materials[i]
			ex := expectedMaterials[i]
			if got != ex {
				t.Errorf("got material %v, expected %v", got, ex)
			}

		}
	}

	data.Set("ean", "invalid")
	req, err = createPostRequest("/materials/add", data)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestNoCacheHandle(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := NoCacheHandle(HomeHandler{})
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	expected := map[string]string{
		"Cache-Control": "no-cache, no-store, must-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}
	for k, v := range expected {
		if rr.Header().Get(k) != v {
			t.Errorf("invalid %v, got %v, expected %v", k, rr.Header().Get(k), v)
		}
	}
}
