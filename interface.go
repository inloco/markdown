package md

// Marshaler is the interface implemented by types that
// can marshal themselves into custom Markdown language.
type Marshaler interface {
	MarshalMarkdown() ([]byte, error)
}
