package recycleme

import (
	"testing"
)

func TestBlacklist(t *testing.T) {
	url := "http://www.upcitemdb.com/upc/3057640136573"
	if err := blacklistDB.Add(url); err != nil {
		t.Fatal(err)
	}
	if ok, err := blacklistDB.Contains(url); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatalf("%v not in blacklist", url)
	}
	_, err := UpcItemDbFetcher.Fetch("3057640136573", blacklistDB)
	if err == nil {
		t.Fatalf("%v not blacklisted", url)
	}
	if err.(*productError).err != errBlacklisted {
		t.Fatalf("not a blacklist error: %v", err)
	}
}

func TestAmazonFetcher(t *testing.T) {
	amazonFetcher, err := newAmazonURLFetcher()

	if amazonFetcher.SecretKey == "" || amazonFetcher.AccessKey == "" || amazonFetcher.AssociateTag == "" {
		t.Log("Missing either AccessKey, SecretKey or AssociateTag. AmazonFetcher will not be tested")
		return
	}
	_, err = amazonFetcher.Fetch("4006381333634", blacklistDB)
	if err != nil {
		if err.(*productError).err != errTooManyProducts {
			t.Fatal(err)
		}
	}
	p, err := amazonFetcher.Fetch("5021991938818", blacklistDB)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Clipper Thé Vert Biologique 20 infusettes" ||
		p.EAN != "5021991938818" ||
		p.URL != "webservices.amazon.fr/5021991938818" ||
		p.ImageURL != "http://ecx.images-amazon.com/images/I/517qE9owUDL.jpg" ||
		p.WebsiteURL != "http://www.amazon.fr/Clipper-Th%C3%A9-Vert-Biologique-infusettes/dp/B011C4L3S0%3FSubscriptionId%3DAKIAIOSACIYSSJVD3IQA%26tag%3Dhowtorecme-21%26linkCode%3Dxm2%26camp%3D2025%26creative%3D165953%26creativeASIN%3DB011C4L3S0" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}
}

func TestDefaultFetchers(t *testing.T) {
	p, err := UpcItemDbFetcher.Fetch("5029053038896", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "Kleenex tissues in a Christmas House box" ||
		p.EAN != "5029053038896" ||
		p.URL != "http://www.upcitemdb.com/upc/5029053038896" ||
		p.ImageURL != "http://www.staples.co.uk/content/images/product/428056_1_xnl.jpg" ||
		p.WebsiteURL != "http://www.upcitemdb.com/upc/5029053038896" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = UpcItemDbFetcher.Fetch("4006381333634", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "Stabilo Boss Original Highlighter Blue" ||
		p.EAN != "4006381333634" ||
		p.URL != "http://www.upcitemdb.com/upc/4006381333634" ||
		p.WebsiteURL != "http://www.upcitemdb.com/upc/4006381333634" ||
		p.ImageURL != "http://ecx.images-amazon.com/images/I/41SfgGjtcpL._SL160_.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = OpenFoodFactsFetcher.Fetch("7613034383808", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "Four à Pierre Royale" || p.EAN != "7613034383808" ||
		p.URL != "http://fr.openfoodfacts.org/api/v0/produit/7613034383808.json" ||
		p.WebsiteURL != "http://fr.openfoodfacts.org/produit/7613034383808/" ||
		p.ImageURL != "http://static.openfoodfacts.org/images/products/761/303/438/3808/front.8.400.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = IsbnSearchFetcher.Fetch("9782501104265", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "le rugby c'est pas sorcier" || p.EAN != "9782501104265" ||
		p.URL != "http://www.isbnsearch.org/isbn/9782501104265" ||
		p.WebsiteURL != "http://www.isbnsearch.org/isbn/9782501104265" ||
		p.ImageURL != "http://ecx.images-amazon.com/images/I/51V4iimUfML._SL194_.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}
}

func TestDefaultFetcher(t *testing.T) {
	fetcher, _ := NewDefaultFetcher()
	_, err := fetcher.Fetch("5029053038896", blacklistDB)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fetcher.Fetch("7613034383808", blacklistDB)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fetcher.Fetch("7640140337517", blacklistDB)
	if err == nil {
		t.Fatal(err)
	}
}
