# Repo practices and principles

## Working principles

- Prefer a self-contained standalone worker repo over sibling-checkout or local `replace` hacks.
- Use `third_party/` as the prefix for copied upstream Stash files.
- Keep sync/update scripts in the repo-root `scripts/` directory and maintain PowerShell support.
- Prefer small, surgical local adapters around vendored upstream code.

## Change hygiene

- Run `go test ./...` before and after changes when code changes are involved.
- Keep generated assets, temp files, and other runtime outputs out of git.
- Update `.agents/references/` when architectural direction or repo workflow meaningfully changes.

## CI expectations

- The manual workflow should be able to refresh vendored files and open a PR automatically.
- CI-friendly changes should not depend on a sibling checkout of the main Stash repository.
