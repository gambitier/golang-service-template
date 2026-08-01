package platform

import (
	domainerr "github.com/gambitier/go-pkgs/errors"
	"github.com/gambitier/go-pkgs/logging"
)

// DomainErrFields extracts structured fields from a domain error for logging.
// Composition lives here so logging and errors packages stay independent.
// Includes cause-chain fields from errors.LogFields while keeping client-safe
// error_msg / error_code for filtering. Always adds error_source when available;
// stack_trace only for INTERNAL errors, filtered to the main module by default.
func DomainErrFields(err error) logging.Fields {
	if err == nil {
		return nil
	}
	fields := logging.Fields{}
	for k, v := range domainerr.LogFields(err) {
		fields[k] = v
	}
	if src := domainerr.OneLineSource(err); src != "" {
		fields["error_source"] = src
	}
	if de, ok := domainerr.As(err); ok {
		fields["error_code"] = string(de.Code)
		fields["error_msg"] = de.Message
		for k, v := range domainerr.CollectFields(err) {
			fields[k] = v
		}
		if de.Code == domainerr.CodeInternal {
			if frames := domainerr.AppStackTraceLines(err, domainerr.StackFrameOptions{}); len(frames) > 0 {
				fields["stack_trace"] = frames
			}
		}
	}
	return fields
}
