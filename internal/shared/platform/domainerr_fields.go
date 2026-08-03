package platform

import (
	domainerr "github.com/gambitier/go-pkgs/errors"
	"github.com/gambitier/go-pkgs/logging"
)

// DomainErrFields adapts errors.LogFields into logging.Fields.
// Enrichment (cause chain, code, message, source, INTERNAL stack) lives in go-pkgs/errors
// so logging and errors stay independent of each other.
func DomainErrFields(err error) logging.Fields {
	return logging.Fields(domainerr.LogFields(err).Map())
}
