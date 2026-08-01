package platform

import (
	"errors"
	"strings"
	"testing"

	domainerr "github.com/gambitier/go-pkgs/errors"
)

func TestDomainErrFields_includesCause(t *testing.T) {
	err := domainerr.Internal("failed to list items", errors.New("context deadline exceeded"), nil)
	fields := DomainErrFields(err)

	msg, _ := fields["error_msg"].(string)
	if msg != "failed to list items" {
		t.Fatalf("error_msg = %q, want opaque domain message", msg)
	}
	code, _ := fields["error_code"].(string)
	if code != string(domainerr.CodeInternal) {
		t.Fatalf("error_code = %q, want INTERNAL", code)
	}
	cause, _ := fields["error_cause"].(string)
	if !strings.Contains(cause, "context deadline exceeded") {
		t.Fatalf("error_cause = %q, want deadline text", cause)
	}
	errorField, _ := fields["error"].(string)
	if !strings.Contains(errorField, "context deadline exceeded") {
		t.Fatalf("error = %q, want cause chain with deadline", errorField)
	}
}

func TestDomainErrFields_internalIncludesSourceAndStack(t *testing.T) {
	err := domainerr.Internal("failed to list items", errors.New("context deadline exceeded"), nil)
	fields := DomainErrFields(err)

	src, _ := fields["error_source"].(string)
	if src == "" || !strings.Contains(src, ".go") {
		t.Fatalf("error_source = %q, want non-empty path with .go", src)
	}
	stack, _ := fields["stack_trace"].(string)
	if strings.TrimSpace(stack) == "" {
		t.Fatal("expected stack_trace for INTERNAL error")
	}
}

func TestDomainErrFields_nonInternalOmitsStack(t *testing.T) {
	err := domainerr.InvalidArgument("name is required", errors.New("empty"), nil)
	fields := DomainErrFields(err)

	if fields["error_code"] != string(domainerr.CodeInvalidArgument) {
		t.Fatalf("error_code = %#v", fields["error_code"])
	}
	if _, ok := fields["stack_trace"]; ok {
		t.Fatalf("did not expect stack_trace for non-INTERNAL, got %#v", fields["stack_trace"])
	}
}

func TestDomainErrFields_plainError(t *testing.T) {
	fields := DomainErrFields(errors.New("db timeout"))
	if fields["error"] != "db timeout" {
		t.Fatalf("error = %#v", fields["error"])
	}
	if _, ok := fields["error_code"]; ok {
		t.Fatalf("did not expect error_code for plain error")
	}
	if _, ok := fields["stack_trace"]; ok {
		t.Fatalf("did not expect stack_trace for plain error")
	}
}
