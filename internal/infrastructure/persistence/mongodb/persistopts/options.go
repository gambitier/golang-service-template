package persistopts

// Options controls repository initialization.
type Options struct {
	// SkipIndexes skips index creation (useful in tests / backfill CLIs).
	SkipIndexes bool
}
