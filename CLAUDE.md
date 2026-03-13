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

# Releases

Releases follow semver and are created by tagging a commit on `main`.

1. Update the version constant in `cmd/tailwind/main.go`:
   ```go
   const version = "X.Y.Z"
   ```
2. Commit the version bump:
   ```
   git add cmd/tailwind/main.go
   git commit -m "release: vX.Y.Z"
   ```
3. Create an annotated tag and push it:
   ```
   git tag -a vX.Y.Z -m "vX.Y.Z"
   git push origin main --follow-tags
   ```

That's it — no CI, no goreleaser, no changelog file. The tag is the release.
