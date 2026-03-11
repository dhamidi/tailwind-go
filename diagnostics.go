package tailwind

import "sort"

// Diagnostics contains runtime statistics about the engine.
type Diagnostics struct {
	TotalCandidates   int      // Number of unique candidates extracted
	MatchedCandidates int      // Number that resolved to utilities
	DroppedCandidates []string // Candidates that didn't match anything
	UtilityCount      int      // Number of registered utility patterns
	VariantCount      int      // Number of registered variants
	ThemeTokenCount   int      // Number of theme tokens
}

// Diagnostics returns runtime statistics about the engine state.
// This calls Flush() to finalize any buffered partial token.
func (e *Engine) Diagnostics() Diagnostics {
	e.Flush()

	e.mu.RLock()
	defer e.mu.RUnlock()

	var dropped []string
	matched := 0

	for c := range e.candidates {
		pc := parseClass(c)
		if pc.ArbitraryProperty {
			matched++
			continue
		}
		utilDef, _ := resolveUtility(pc, e.utilIndex)
		if utilDef != nil {
			matched++
		} else {
			dropped = append(dropped, c)
		}
	}

	sort.Strings(dropped) // deterministic output

	return Diagnostics{
		TotalCandidates:   len(e.candidates),
		MatchedCandidates: matched,
		DroppedCandidates: dropped,
		UtilityCount:      len(e.utilIndex.static) + len(e.utilIndex.dynamic),
		VariantCount:      len(e.variants),
		ThemeTokenCount:   len(e.theme.Tokens),
	}
}
