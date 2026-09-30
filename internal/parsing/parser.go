package parsing

// Parser extracts structured fields from a raw log string of a known format.
type Parser interface {
	Name() string
	Parse(raw string) (ParsedEvent, error)
}
