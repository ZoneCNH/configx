package configx

import "fmt"

// MergeStrategy controls how conflicting keys from multiple sources are resolved.
type MergeStrategy int

const (
	// LastWins overwrites earlier values with later ones.
	LastWins MergeStrategy = iota
	// FirstWins keeps the first value seen and ignores later ones.
	FirstWins
	// ErrorOnConflict returns an error when a key appears in multiple sources.
	ErrorOnConflict

	// MergeLastWins is an alias for LastWins.
	MergeLastWins = LastWins
	// MergeFirstWins is an alias for FirstWins.
	MergeFirstWins = FirstWins
	// MergeErrorOnConflict is an alias for ErrorOnConflict.
	MergeErrorOnConflict = ErrorOnConflict
)

func mergeValue(values Map, key string, value Value, strategy MergeStrategy) error {
	prev, exists := values[key]
	if !exists {
		values[key] = value
		return nil
	}
	switch strategy {
	case LastWins:
		prev.Overridden = true
		values[key] = prev
		values[key] = value
	case FirstWins:
		return nil
	case ErrorOnConflict:
		return fmt.Errorf("merge conflict for key %q", key)
	default:
		return fmt.Errorf("unknown merge strategy %d", strategy)
	}
	return nil
}
