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
		p.ImageURL != "http://static.openfoodfacts.org/images/products/761/303/438/3808/front_fr.8.400.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = IGalerieFetcher.Fetch("8714789941011", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "" || p.EAN != "8714789941011" ||
		p.URL != "http://90.80.54.225/?search=8714789941011" ||
		p.WebsiteURL != "http://90.80.54.225/?search=8714789941011" ||
		p.ImageURL != "http://90.80.54.225/albums/Permanent/LOGIDIS2/8714789941011.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = localProductDB.Fetch("ean_test", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "name_test" || p.EAN != "ean_test" ||
		p.URL != "" ||
		p.WebsiteURL != "" || p.WebsiteName != "website_name_test" ||
		p.ImageURL != "" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = StarrymartFetcher.Fetch("4897878100026", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "Nissin Cup Noodles Beef Flavour 75g" || p.EAN != "4897878100026" ||
		p.URL != "https://starrymart.co.uk/catalogsearch/result/?q=4897878100026" ||
		p.WebsiteURL != "https://starrymart.co.uk/nissin-cup-noodle-beef.html" ||
		p.ImageURL != "https://starrymart.co.uk/media/catalog/product/cache/1/small_image/202x303/9df78eab33525d08d6e5fb8d27136e95/4/8/4897878100026_5.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = MisterPharmaWebFetcher.Fetch("3400937688369", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "HUMEX ALLERGIE CETIRIZINE 10 mg, comprimé pelliculé sécable" || p.EAN != "3400937688369" ||
		p.URL != "http://www.misterpharmaweb.com/recherche-resultats.php?search_in_description=1&ac_keywords=3400937688369" ||
		p.WebsiteURL != "http://www.misterpharmaweb.com/humex-allergie-cetirizine-10-mg-comprime-pellicule-secable-xml-351_365-2807.html" ||
		p.ImageURL != "http://www.misterpharmaweb.com/images/imagecache/cetir_1425058129_180x180.jpg" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = MedisparFetcher.Fetch("3400936864986", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "NUROFEN 400mg CPR ENR B/12" || p.EAN != "3400936864986" ||
		p.URL != "http://www.meddispar.fr/content/search?search_by_name=&search_by_cip=3400936864986" ||
		p.WebsiteURL != "http://www.meddispar.fr/Medicaments/NUROFEN-400-B-12/(type)/cip/(value)/3400936864986" ||
		p.ImageURL != "" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}

	p, err = PicardFetcher.Fetch("3270160891382", blacklistDB)
	if err != nil {
		t.Error(err)
	} else if p.Name != "2 quiches lorraines" || p.EAN != "3270160891382" ||
		p.URL != "http://www.picard.fr/recherche?q=3270160891382" ||
		p.WebsiteURL != "http://www.picard.fr/produits/2-quiches-lorraines-000000000000089138.html" ||
		p.ImageURL != "http://demandware.edgesuite.net/sits_pod39/dw/image/v2/AAHV_PRD/on/demandware.static/-/Sites-catalog-picard/default/dwe4154de5/produits/entrees-tartes-salades/pack/000000000000089138_P.png?sw=140&sh=82" {
		t.Errorf("Some attributes are invalid for: %v", p)
	}
}

func TestLocalProductDB(t *testing.T) {
	if localProductDB.IsURLValidForEAN("/", "EAN") {
		t.Error("/ should be false")
	}
	if !localProductDB.IsURLValidForEAN("/local/website_name_test/ean_test", "ean_test") {
		t.Error("/local/website_name_test/ean_test should be true")
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

	// Fake EAN, should return an error
	_, err = fetcher.Fetch("4012345123456", blacklistDB)
	if err == nil {
		t.Fatal("Error should not be nil")
	}
}
