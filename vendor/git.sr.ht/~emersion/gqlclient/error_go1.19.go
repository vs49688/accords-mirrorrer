//go:build !go1.20

package gqlclient

func joinErrors(errs []Error) error {
	return &errs[0]
}
