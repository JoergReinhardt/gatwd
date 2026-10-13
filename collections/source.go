package collections

import "strconv"

// Source says where a value came from: a file, a URL, a rendered stream — whatever
// the producer calls its origin — and, when the origin holds several values, which one.
//
// # Why this exists
//
// When ingest goes wrong, the question a person asks first is "which file?". The error
// that reports two records colliding on one identity (DuplicateNodeError) can quote what
// the records contain, but not where they were read from, because a record is just its
// content: two files that hold the same text yield two equal values. Source is the
// provenance a producer (an adapter reading files) attaches to the value it yields, so an
// error about the value can name its origin.
//
// # How it is carried — and what it never changes
//
// The pattern is the metadata of Clojure's `with-meta`: a *side channel* on the value.
// A HAMT (the shape of a record) carries one optional Source (HAMT.WithSource,
// HAMT.Source). It is invisible to everything that reads the value: Get, Seq, Keys,
// Values, Len, Equal, JSON and the DSL's printing and comparison never see it, so
// a record read from a file and the same record fetched from elsewhere are equal and
// behave identically in every rule and relation. Only the error paths that ask
// (SourceOf) show it. A producer with no useful origin (a live API) attaches nothing,
// and an error about its values reads exactly as before.
//
// Source is deliberately domain-free: Origin is an opaque string.
type Source struct {
	// Origin names where the value was read: a file path, a URL, "kustomize build DIR".
	Origin string
	// Document is the 1-based position of the value inside Origin when Origin holds several
	// (the second `---`-separated document of a YAML file); 0 means the origin holds one.
	Document int
}

// IsZero reports whether s names no origin.
func (s Source) IsZero() bool { return s.Origin == "" && s.Document == 0 }

// String renders the origin and, for a multi-value origin, the position:
// "deploy/web.yaml" or "deploy/all.yaml, document 3".
func (s Source) String() string {
	if s.Document > 0 {
		return s.Origin + ", document " + strconv.Itoa(s.Document)
	}
	return s.Origin
}

// Sourced is implemented by values that remember their Source (HAMT does).
type Sourced interface {
	// Source returns the attached Source and whether there is one.
	Source() (Source, bool)
}

// SourceOf returns the Source v carries, if v is Sourced and has one.
//
//	if src, ok := collections.SourceOf(payload); ok {
//	    fmt.Println("read from", src)
//	}
func SourceOf(v any) (Source, bool) {
	if s, ok := v.(Sourced); ok {
		return s.Source()
	}
	return Source{}, false
}
