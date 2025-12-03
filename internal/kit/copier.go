package kit

import "github.com/jinzhu/copier"

func AutoCopy[S, T any](toValue *T, fromValue S) (*T, error) {
	if err := copier.Copy(toValue, fromValue); err != nil {
		return nil, err
	}
	return toValue, nil
}
