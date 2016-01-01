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

	p, err = OpenFoodFactsFetcher.Fetch("7613034383808")
	if err != nil {
		t.Error(err)
	} else if p.Name != "Four à Pierre Royale" || p.EAN != "7613034383808" ||
		p.URL != "http://fr.openfoodfacts.org/api/v0/produit/7613034383808.json" ||
		p.ImageURL != "http://static.openfoodfacts.org/images/products/761/303/438/3808/front.8.400.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}
}

func TestScrap(t *testing.T) {
	product, err := Scrap("5029053038896")
	if err != nil {
		t.Error(err)
	}
	if product == nil {
		t.Errorf("product should not be nil")
	}

	product, err = Scrap("7613034383808")
	if err != nil {
		t.Error(err)
	}
	if product == nil {
		t.Errorf("product should not be nil")
	}
	product, err = Scrap("9782123456803")
	if product != nil {
		t.Errorf("product should be nil: %v", product)
	}
	if err == nil {
		t.Error(err)
	}
}
