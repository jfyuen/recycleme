package recycleme

import "fmt"

var errNotFound = fmt.Errorf("product not found")
var errBlacklisted = fmt.Errorf("product blacklisted for url")
var errTooManyProducts = fmt.Errorf("too many products found")

type ProductError struct {
	EAN, URL string
	err      error
}

func (err ProductError) Error() string {
	return fmt.Sprintf("%v for %v at %v", err.err, err.EAN, err.URL)
}

func NewProductError(ean, url string, err error) *ProductError {
	return &ProductError{EAN: ean, URL: url, err: err}
}
