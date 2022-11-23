package index

import "golang.org/x/xerrors"

var (
	ErrNotFound      = xerrors.New("not found")
	ErrMissingLinkID = xerrors.New("document does not provide a valid link ID")
)
