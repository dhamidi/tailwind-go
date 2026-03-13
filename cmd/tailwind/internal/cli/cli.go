// Package cli provides a convenience layer on top of Go's flag package
// that supports short and long flag aliases, optional-value flags, and
// formatted help output matching the official tailwindcss CLI style.
package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrHelp is returned by Parse when --help or -h is encountered.
var ErrHelp = errors.New("help requested")

type flagKind int

const (
	kindString flagKind = iota
	kindBool
	kindOptional
)

type flagDef struct {
	short        string
	long         string
	description  string
	defaultValue string
	kind         flagKind
	value        *string
	boolPtr      *bool
}

// Command represents a CLI command with flags and a run function.
type Command struct {
	name        string
	description string
	flags       []*flagDef
	// maps short and long names to the flag definition
	lookup map[string]*flagDef
}

// NewCommand creates a new command with the given name and description.
func NewCommand(name, description string) *Command {
	return &Command{
		name:        name,
		description: description,
		lookup:      make(map[string]*flagDef),
	}
}

// StringFlag registers a string flag with short and long names.
func (c *Command) StringFlag(short, long, defaultValue, description string) *string {
	v := new(string)
	*v = defaultValue
	f := &flagDef{
		short:        short,
		long:         long,
		description:  description,
		defaultValue: defaultValue,
		kind:         kindString,
		value:        v,
	}
	c.register(f)
	return v
}

// BoolFlag registers a boolean flag with short and long names.
// The returned pointer holds "true" or "false" as a string, but the
// public API returns *bool for convenience.  Internally we store
// booleans as string pointers so that all flags share one backing type.
func (c *Command) BoolFlag(short, long string, defaultValue bool, description string) *bool {
	result := new(bool)
	*result = defaultValue
	def := ""
	if defaultValue {
		def = "true"
	}
	v := new(string)
	*v = def
	f := &flagDef{
		short:        short,
		long:         long,
		description:  description,
		defaultValue: def,
		kind:         kindBool,
		value:        v,
	}
	c.register(f)
	// We keep the bool pointer in sync after parse via a small wrapper.
	// To do that without extra complexity we store a reference and
	// resolve it in Parse.
	f.boolPtr = result
	return result
}

// OptionalFlag registers a flag that can be used as a boolean or with a value.
// --watch (no value) sets the pointer to "" (truthy, default mode).
// --watch=always sets the pointer to "always".
// When not provided at all the pointer holds defaultValue.
func (c *Command) OptionalFlag(short, long, defaultValue, description string) *string {
	v := new(string)
	*v = defaultValue
	f := &flagDef{
		short:        short,
		long:         long,
		description:  description,
		defaultValue: defaultValue,
		kind:         kindOptional,
		value:        v,
	}
	c.register(f)
	return v
}

func (c *Command) register(f *flagDef) {
	c.flags = append(c.flags, f)
	if f.short != "" {
		c.lookup[f.short] = f
	}
	if f.long != "" {
		c.lookup[f.long] = f
	}
}

// Parse parses the given argument list (typically os.Args[1:]).
func (c *Command) Parse(args []string) error {
	i := 0
	for i < len(args) {
		arg := args[i]

		// Handle --flag=value
		if strings.HasPrefix(arg, "--") {
			name, val, hasEq := strings.Cut(arg[2:], "=")
			if name == "help" {
				return ErrHelp
			}
			f, ok := c.lookup[name]
			if !ok {
				return fmt.Errorf("unknown flag: --%s", name)
			}
			if hasEq {
				*f.value = val
				if f.kind == kindBool && f.boolPtr != nil {
					*f.boolPtr = val == "true" || val == "1"
				}
			} else {
				switch f.kind {
				case kindBool:
					*f.value = "true"
					if f.boolPtr != nil {
						*f.boolPtr = true
					}
				case kindOptional:
					*f.value = ""
				case kindString:
					i++
					if i >= len(args) {
						return fmt.Errorf("flag --%s requires a value", name)
					}
					*f.value = args[i]
				}
			}
			i++
			continue
		}

		// Handle -x or -x=value or -x value
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			name, val, hasEq := strings.Cut(arg[1:], "=")
			if name == "h" {
				return ErrHelp
			}
			f, ok := c.lookup[name]
			if !ok {
				return fmt.Errorf("unknown flag: -%s", name)
			}
			if hasEq {
				*f.value = val
				if f.kind == kindBool && f.boolPtr != nil {
					*f.boolPtr = val == "true" || val == "1"
				}
			} else {
				switch f.kind {
				case kindBool:
					*f.value = "true"
					if f.boolPtr != nil {
						*f.boolPtr = true
					}
				case kindOptional:
					*f.value = ""
				case kindString:
					i++
					if i >= len(args) {
						return fmt.Errorf("flag -%s requires a value", name)
					}
					*f.value = args[i]
				}
			}
			i++
			continue
		}

		// Non-flag argument: stop parsing (positional args not supported).
		return fmt.Errorf("unexpected argument: %s", arg)
	}

	// Sync bool pointers.
	for _, f := range c.flags {
		if f.kind == kindBool && f.boolPtr != nil {
			*f.boolPtr = *f.value == "true"
		}
	}

	return nil
}

// PrintHelp writes formatted help to the given writer.
func (c *Command) PrintHelp(w io.Writer) {
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  %s [options…]\n", c.name)
	fmt.Fprintf(w, "\nOptions:\n")

	type entry struct {
		left  string
		right string
	}

	var entries []entry
	maxLeft := 0

	for _, f := range c.flags {
		var left string
		if f.short != "" && f.long != "" {
			left = fmt.Sprintf("-%s, --%s", f.short, f.long)
		} else if f.long != "" {
			left = fmt.Sprintf("     --%s", f.long)
		} else {
			left = fmt.Sprintf("-%s", f.short)
		}

		desc := f.description
		if f.defaultValue != "" {
			desc += fmt.Sprintf(" [default: `%s`]", f.defaultValue)
		}

		entries = append(entries, entry{left: left, right: desc})
		if len(left) > maxLeft {
			maxLeft = len(left)
		}
	}

	for _, e := range entries {
		padded := e.left + strings.Repeat(" ", maxLeft-len(e.left))
		fmt.Fprintf(w, "  %s ·· %s\n", padded, e.right)
	}
}
