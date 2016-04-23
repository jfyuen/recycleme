package recycleme

import (
	"errors"
	"fmt"
)

var errNotFound = fmt.Errorf("product not found")
var errInvalidEAN = fmt.Errorf("invalid ean")
var errBlacklisted = fmt.Errorf("product blacklisted for url")
var errTooManyProducts = fmt.Errorf("too many products found")
var errPackageNotFound = errors.New("ean not found in packages db")

type productError struct {
	EAN, URL string
	err      error
}

func (err productError) Error() string {
	return fmt.Sprintf("%v for %v at %v", err.err, err.EAN, err.URL)
}

func newProductError(ean, url string, err error) *productError {
	return &productError{EAN: ean, URL: url, err: err}
}
