//go:build go1.20

package gqlclient

import (
	"errors"
)

func joinErrors(errs []Error) error {
	l := make([]error, len(errs))
	for i := range errs {
		l[i] = &errs[i]
	}
	return errors.Join(l...)
}
