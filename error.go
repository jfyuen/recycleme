package recycleme

import "fmt"

var errNotFound = fmt.Errorf("product not found")
var errTooManyProducts = fmt.Errorf("too many products found")

type ProductError struct {
	EAN, URL string
	msg      string
}

func (err ProductError) Error() string {
	return fmt.Sprintf("%v for %v at %v", err.msg, err.EAN, err.URL)
}

func NewProductError(ean, url string, err error) *ProductError {
	return &ProductError{EAN: ean, URL: url, msg: err.Error()}
}
