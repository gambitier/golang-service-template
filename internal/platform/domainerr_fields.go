package platform

import (
	"github.com/gambitier/go-pkgs/errors/domainerr"
	"github.com/gambitier/go-pkgs/logging"
)

// DomainErrFields extracts structured fields from a domain error for logging.
// Composition lives here so logging and errors packages stay independent.
func DomainErrFields(err error) logging.Fields {
	if err == nil {
		return nil
	}
	de, ok := domainerr.As(err)
	if !ok {
		return logging.Fields{"error": err.Error()}
	}
	fields := logging.Fields{
		"error":       de.Error(),
		"error_code":  string(de.Code),
		"error_msg":   de.Message,
	}
	collected := domainerr.CollectFields(err)
	for k, v := range collected {
		fields[k] = v
	}
	return fields
}
