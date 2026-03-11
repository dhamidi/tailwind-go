# tailwind-go

This is pure-Go implementation of TailwindCSS.

The normative source is spec.md

# Rules

- no dependencies outside of the standard library,
- unit tests directly in this package,
- unit tests must not do I/O: such operations must be expressed in terms of interfaces like io.Writer/io.Reader
