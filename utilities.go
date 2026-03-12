package tailwind

import "strings"

// CompileFn is a function that generates CSS declarations for a utility candidate.
// Returns nil if the candidate doesn't match.
type CompileFn func(candidate ResolvedCandidate) []Declaration

// ResolvedCandidate contains the parsed class information passed to a compile function.
type ResolvedCandidate struct {
	Value     string       // e.g., "4", "blue-500", "" (for static)
	Arbitrary string       // e.g., "#ff0000" for bg-[#ff0000]
	Modifier  string       // e.g., "50" for bg-red-500/50
	Negative  bool         // true for -translate-x-4
	TypeHint  string       // e.g., "color" for bg-[color:red]
	Fraction  string       // e.g., "1/2" for w-1/2
	Theme     *ThemeConfig // theme configuration for token resolution
}

// UtilityRegistration represents a programmatically registered utility.
type UtilityRegistration struct {
	Name      string    // e.g., "flex", "bg", "p"
	Kind      string    // "static" or "functional"
	CompileFn CompileFn // generates declarations for a candidate
	Order     int       // definition order for stable CSS output
	Selector  string    // optional child selector suffix
}

func (u *UtilityRegistration) utilityPattern() string  { return u.Name }
func (u *UtilityRegistration) utilityIsStatic() bool   { return u.Kind == "static" }
func (u *UtilityRegistration) utilityOrder() int       { return u.Order }
func (u *UtilityRegistration) utilitySelector() string { return u.Selector }

// staticUtility creates a UtilityRegistration for a fixed-declaration utility
// (e.g., flex, block, sr-only). The declarations are emitted as-is.
func staticUtility(name string, decls []Declaration) *UtilityRegistration {
	return &UtilityRegistration{
		Name: name,
		Kind: "static",
		CompileFn: func(candidate ResolvedCandidate) []Declaration {
			if candidate.Negative {
				var negated []Declaration
				for _, d := range decls {
					neg := negateValue(d.Value)
					if neg == "" {
						return nil
					}
					negated = append(negated, Declaration{Property: d.Property, Value: neg})
				}
				return negated
			}
			return decls
		},
	}
}

// functionalUtility creates a UtilityRegistration for a theme-resolved utility
// (e.g., z-*, order-*) that accepts dynamic values with type inference.
func functionalUtility(name string, compileFn CompileFn) *UtilityRegistration {
	return &UtilityRegistration{
		Name:      name,
		Kind:      "functional",
		CompileFn: compileFn,
	}
}

// spacingUtility creates a UtilityRegistration for a spacing-based utility
// (e.g., p-*, m-*, gap-*) with multiplier support.
func spacingUtility(name string, compileFn CompileFn) *UtilityRegistration {
	return &UtilityRegistration{
		Name:      name,
		Kind:      "functional",
		CompileFn: compileFn,
	}
}

// colorUtility creates a UtilityRegistration for a color utility
// (e.g., bg-*, text-*, border-*) with opacity modifier support.
func colorUtility(name string, compileFn CompileFn) *UtilityRegistration {
	return &UtilityRegistration{
		Name:      name,
		Kind:      "functional",
		CompileFn: compileFn,
	}
}

// resolveEntryDeclarations resolves declarations for a utilityEntry,
// dispatching to CompileFn for Go-registered utilities or to
// resolveDeclarations for CSS-parsed utilities.
func resolveEntryDeclarations(entry utilityEntry, valueStr string, pc ParsedClass, theme *ThemeConfig) []Declaration {
	switch u := entry.(type) {
	case *UtilityRegistration:
		candidate := ResolvedCandidate{
			Value:     valueStr,
			Arbitrary: pc.Arbitrary,
			Modifier:  pc.Modifier,
			Negative:  pc.Negative,
			TypeHint:  pc.TypeHint,
			Theme:     theme,
		}
		if strings.Contains(valueStr, "/") {
			candidate.Fraction = valueStr
		}
		return u.CompileFn(candidate)
	case *UtilityDef:
		return resolveDeclarations(u, valueStr, pc, theme)
	}
	return nil
}

// registerGoUtilities registers proof-of-concept Go-based static utilities.
// These replace the equivalent CSS-parsed definitions and produce identical output.
func registerGoUtilities(idx *utilityIndex) {
	register := func(reg *UtilityRegistration) {
		// Inherit order from existing CSS-parsed utility if present.
		if reg.Kind == "static" {
			if existing, ok := idx.static[reg.Name]; ok {
				reg.Order = existing.utilityOrder()
			}
		} else {
			for _, entry := range idx.dynamic {
				if entry.utilityPattern() == reg.Name {
					reg.Order = entry.utilityOrder()
					break
				}
			}
		}
		idx.add(reg)
	}

	register(staticUtility("flex", []Declaration{{Property: "display", Value: "flex"}}))
	register(staticUtility("block", []Declaration{{Property: "display", Value: "block"}}))
	register(staticUtility("hidden", []Declaration{{Property: "display", Value: "none"}}))
}
