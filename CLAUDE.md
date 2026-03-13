# tailwind-go

This is pure-Go implementation of TailwindCSS.

The normative source is spec.md

# Rules

- no dependencies outside of the standard library,
- unit tests directly in this package,
- unit tests must not do I/O: such operations must be expressed in terms of interfaces like io.Writer/io.Reader

# Driver programs

- Driver programs under `cmd/` (other than the main `cmd/tailwind` CLI) are internal, throwaway tools for local development and experimentation.
- They should NOT be committed to the repository. Add them to `.gitignore` or keep them local.
- Any functionality worth preserving from a driver program should be captured as a unit test instead.
