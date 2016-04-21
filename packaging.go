package recycleme

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

var ErrPackageNotFound = errors.New("ean not found in packages db")

type Bin struct {
	Name string `json:"name" bson:"name"`
	ID   uint   `json:"id" bson:"_id,omitempty"`
}

// A Material composes Packaging, different Materials go to different Bin, event ones that may be close enough
// For example, in Paris, plastic bags go to the green bin, but plastic bottles go to the yellow bin
type Material struct {
	ID   uint   `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
}

// Packaging related to a Product
// Several products may have the same types of packaging
// For example, a pizza box and a frozen product may both have a cardboard box and a plastic foil
type Package struct {
	EAN       string
	Materials []Material
}

func (p Package) String() string {
	s := make([]string, len(p.Materials), len(p.Materials))
	for i, m := range p.Materials {
		s[i] = m.Name
	}
	return fmt.Sprintf("Product %v is composed of %v", p.EAN, strings.Join(s, ", "))
}

type PackagesDB interface {
	Get(ean string) (Package, error)
	GetBins([]Material) (map[Material]Bin, error)
	Set(ean string, m []Material) error
}

func withMgoSession(s *mgo.Session, fn func(s *mgo.Session) error) error {
	session := s.Copy()
	defer session.Close()
	return fn(s)
}

type mgoDB struct {
	session *mgo.Session
}
type MaterialDB interface {
	GetAll() ([]Material, error)
}

func (db mgoPackagesDB) GetAll() ([]Material, error) {
	var m []Material
	err := withMgoSession(db.session, func(s *mgo.Session) error {
		db := s.DB("")
		collection := db.C("materials")
		if err := collection.Find(nil).All(&m); err != nil {
			return err
		}

		return nil
	})
	return m, err
}

type mgoPackagesDB struct {
	mgoDB
	packagesColName, materialsColName, binsColName, materialsToBinsColName string
}

type mgoPackageItem struct {
	EAN         string `json:"ean" bson:"ean"`
	MaterialIDs []uint `json:"material_ids" bson:"material_ids"`
}

func NewMgoPackageDB(s *mgo.Session, colPrefix string) *mgoPackagesDB {
	return &mgoPackagesDB{mgoDB: mgoDB{session: s},
		packagesColName:        colPrefix + "packages",
		materialsColName:       colPrefix + "materials",
		binsColName:            colPrefix + "bins",
		materialsToBinsColName: colPrefix + "materials_to_bins",
	}
}

func (db mgoPackagesDB) Get(ean string) (Package, error) {
	p := Package{EAN: ean}
	err := withMgoSession(db.session, func(s *mgo.Session) error {
		db_ := s.DB("")
		collection := db_.C(db.packagesColName)
		item := mgoPackageItem{}
		if err := collection.Find(bson.M{"ean": ean}).One(&item); err != nil {
			if err == mgo.ErrNotFound {
				return ErrPackageNotFound
			}
			return err
		}
		materialsCol := db_.C(db.materialsColName)
		req := bson.M{"_id": bson.M{"$in": item.MaterialIDs}}
		if err := materialsCol.Find(req).All(&p.Materials); err != nil {
			return err
		}

		return nil
	})
	return p, err
}

func (db mgoPackagesDB) Set(ean string, m []Material) error {
	return withMgoSession(db.session, func(s *mgo.Session) error {
		item := mgoPackageItem{EAN: ean}
		materialIDSet := make(map[uint]struct{})
		for _, material := range m {
			if _, ok := materialIDSet[material.ID]; ok {
				continue
			}
			materialIDSet[material.ID] = struct{}{}
			item.MaterialIDs = append(item.MaterialIDs, material.ID)
		}
		collection := s.DB("").C(db.packagesColName)
		if _, err := collection.Upsert(bson.M{"ean": ean}, item); err != nil {
			return err
		}
		return nil
	})
}

func (db mgoPackagesDB) GetBins(m []Material) (map[Material]Bin, error) {
	r := make(map[Material]Bin)
	return r, withMgoSession(db.session, func(s *mgo.Session) error {
		mIDs := make([]uint, len(m), len(m))
		mIDMap := make(map[uint]Material)
		for _, material := range m {
			mIDs = append(mIDs, material.ID)
			mIDMap[material.ID] = material
		}
		collection := s.DB("").C(db.materialsToBinsColName)
		var mIDsToBinIDs []struct {
			MaterialID uint `bson:"material_id"`
			BinID      uint `bson:"bin_id"`
		}
		if err := collection.Find(bson.M{"material_id": bson.M{"$in": mIDs}}).All(&mIDsToBinIDs); err != nil {
			return err
		}
		var binIDs []uint
		for _, mb := range mIDsToBinIDs {
			binIDs = append(binIDs, mb.BinID)
		}
		collection = s.DB("").C(db.binsColName)
		var bins []Bin
		if err := collection.Find(bson.M{"_id": bson.M{"$in": binIDs}}).All(&bins); err != nil {
			return err
		}

		binMap := make(map[uint]Bin)
		for _, b := range bins {
			binMap[b.ID] = b
		}

		for _, mIDToBinID := range mIDsToBinIDs {
			material := mIDMap[mIDToBinID.MaterialID]
			r[material] = binMap[mIDToBinID.BinID]
		}

		return nil
	})
}

// ProductPackage links a Product and its packages
type ProductPackage struct {
	Product   `json:",inline"`
	Materials []Material `json:"materials"`
}

func NewProductPackage(p Product, db PackagesDB) (ProductPackage, error) {
	pp := ProductPackage{Product: p}
	pkg, err := db.Get(p.EAN)
	if err != nil {
		if err == ErrPackageNotFound {
			pp.Materials = make([]Material, 0, 0)
			return pp, nil
		}
		return pp, err
	}
	pp.Materials = pkg.Materials
	return pp, nil
}

func (pp ProductPackage) ThrowAway(db PackagesDB) (map[Material]Bin, error) {
	return db.GetBins(pp.Materials)
}

type throwAwaypackage struct {
	Product   ProductPackage    `json:"product"`
	ThrowAway map[string]string `json:"throwAway"`
}

func (pp ProductPackage) ThrowAwayJSON(db PackagesDB) ([]byte, error) {
	throwAway, err := pp.ThrowAway(db)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for k, v := range throwAway {
		out[k.Name] = v.Name
	}
	return json.Marshal(throwAwaypackage{pp, out})
}
