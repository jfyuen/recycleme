package recycleme

import (
	"testing"
)

func TestDefaultFetchers(t *testing.T) {
	p, err := UpcItemDbFetcher.Fetch("5029053038896")
	if err != nil {
		t.Error(err)
	} else if p.Name != "Kleenex tissues in a Christmas House box" ||
		p.EAN != "5029053038896" ||
		p.URL != "http://www.upcitemdb.com/upc/5029053038896" ||
		p.ImageURL != "http://www.staples.co.uk/content/images/product/428056_1_xnl.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = UpcItemDbFetcher.Fetch("4006381333634")
	if err != nil {
		t.Error(err)
	} else if p.Name != "Stabilo Boss Original Highlighter Blue" ||
		p.EAN != "4006381333634" ||
		p.URL != "http://www.upcitemdb.com/upc/4006381333634" ||
		p.ImageURL != "http://ecx.images-amazon.com/images/I/41SfgGjtcpL._SL160_.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = OpenFoodFactsFetcher.Fetch("7613034383808")
	if err != nil {
		t.Error(err)
	} else if p.Name != "Four à Pierre Royale" || p.EAN != "7613034383808" ||
		p.URL != "http://fr.openfoodfacts.org/api/v0/produit/7613034383808.json" ||
		p.ImageURL != "http://static.openfoodfacts.org/images/products/761/303/438/3808/front.8.400.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = IsbnSearchFetcher.Fetch("9782501104265")
	if err != nil {
		t.Error(err)
	} else if p.Name != "le rugby c'est pas sorcier" || p.EAN != "9782501104265" ||
		p.URL != "http://www.isbnsearch.org/isbn/9782501104265" ||
		p.ImageURL != "http://ecx.images-amazon.com/images/I/51V4iimUfML._SL194_.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}
}

func TestScrap(t *testing.T) {
	_, err := Scrap("5029053038896")
	if err != nil {
		t.Error(err)
	} else {
		_, err = Scrap("7613034383808")
		if err != nil {
			t.Error(err)
		} else {
			_, err = Scrap("9782123456803")
			if err == nil {
				t.Error(err)
			}
		}
	}
}
