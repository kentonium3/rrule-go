# Contributing

Thanks for helping. This is a maintained fork of teambition/rrule-go.

## Reporting a problem

Open an issue using one of the forms. For a bug, give the exact rule and DTSTART, what you expected
and what you got. Behaviour follows [RFC 5545](https://www.rfc-editor.org/rfc/rfc5545), so if you can,
cite the section your expectation rests on. What python-dateutil returns for the same rule is useful
too, but where it differs from the RFC, the RFC wins.

Security problems (panics, hangs, runaway resource use) go through private reporting instead. See
[SECURITY.md](SECURITY.md).

## Changing the code

- Keep the module path (`github.com/teambition/rrule-go`), the public API and the Go 1.16 minimum.
  Changes that would break existing callers need an issue first.
- Add a test that fails before your change and passes after it.
- Before opening a pull request, run:

  ```bash
  go vet ./...
  go test -race ./...
  ```

- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages, for example
  `fix: stop BYSETPOS from skipping the last period`.
- Open pull requests against the `main` branch.

## Changelog

Include a consumer-focused entry in [CHANGELOG.md](CHANGELOG.md) for user-facing
changes, or explain in the pull request why no entry is needed. Add it under
`## [Unreleased]`, using only populated categories in this order: Breaking,
Upgrade Notes, Added, Changed, Fixed, Deprecated, Removed, Security.

Start with a short bold headline describing the effect. Explain the previous and
new behavior when that helps users; breaking changes must include migration
instructions. Link public issues or pull requests and credit retained contributions.
Routine internal changes need no entry. Keep future plans in the roadmap and
private security details in private reporting until disclosure is appropriate.

### Preparing a release

Review the changes included in the release, move the relevant Unreleased entries
into `## [X.Y.Z] - YYYY-MM-DD`, and leave `## [Unreleased]` for subsequent work.
Use the actual release date and exact version of the intended Git tag (for
example, `v1.8.4` corresponds to `[1.8.4]`). The `go 1.16` directive in `go.mod`
is the language compatibility floor, not the release version.

Use that version section as the GitHub release notes. Review it alongside the
code and migration guidance, and pass the existing code gates before publishing.
Released sections are historical records: correct factual errors when necessary,
but do not restyle them or add changes that were not in that release.

By contributing you agree that your work is released under the MIT license in [LICENSE](LICENSE).
