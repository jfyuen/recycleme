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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.New(ioutil.Discard, "", 0)
		AddBlacklistHandler(Blacklist, w, r, logger, nopFetcher, m.sendMail)
	})

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

	if ok, err := Blacklist.Contains(url); err != nil {
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

	nopFetcher := FetchableURL{URL: "http://www.example.com/%s/", WebsiteName: "Example.com"}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ThrowAwayHandler(Packages, w, r, nopFetcher)
	})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	url := fullURL(nopFetcher.URL, ean)
	expected := fmt.Sprintf(`{"product":{"ean":"%s","name":"TEST","url":"%s","imageURL":"","websiteURL":"","websiteName":"%s","materials":[]},"throwAway":{}}`, ean, url, nopFetcher.WebsiteName)
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

func TestAddPackageHandler(t *testing.T) {
	data := url.Values{}
	ean := "5021991938818"
	data.Set("ean", ean)
	expectedMaterials := []Material{Material{ID: 0, Name: "m0"}, Material{ID: 1, Name: "m1"}}
	materialsJSON, err := json.Marshal(expectedMaterials)
	if err != nil {
		t.Fatal(err)
	}
	data.Set("materials", string(materialsJSON))

	req, err := createPostRequest("/materials/add", data)
	if err != nil {
		t.Fatal(err)
	}
	m := newMailTester("Adding package for "+ean, fmt.Sprintf("Materials added to %v:\n%v", ean, "[{0 m0} {1 m1}]"))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.New(ioutil.Discard, "", 0)
		AddPackageHandler(Packages, w, r, logger, m.sendMail)
	})

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

	if v, err := Packages.Get(ean); err != nil {
		if err == ErrPackageNotFound {
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
