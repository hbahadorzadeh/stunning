# Contributing to Stunning

Thanks for your interest in improving Stunning! Contributions of all kinds are
welcome — code, docs, tests, bug reports, and plugin ideas.

## Contributor License Agreement (required)

Before any pull request can be merged, you must agree to the
[Contributor License Agreement (CLA)](CLA.md).

**By opening a pull request, you confirm that you have read and agree to the
CLA.** It grants the project owner a license to your contribution (including the
right to relicense), while you keep ownership of your work. This keeps the project
legally clean and lets it be licensed freely in the future.

If you contribute on behalf of a **company**, an authorized representative must
email **h.bahadorzadeh@gmail.com** to confirm the entity agrees to the CLA and to
list the authorized GitHub accounts.

> Tip: the project enforces this automatically with the
> [CLA Assistant](https://github.com/contributor-assistant/github-action) GitHub
> Action, which posts a sign-off checkbox on every PR.

## Development workflow

1. Fork the repository and create a feature branch.
2. Make your changes, with tests where it makes sense.
3. Run the checks locally:
   ```bash
   go build ./...
   go test -race ./...
   go vet ./...
   ```
4. Keep commits focused and write clear messages.
5. Open a pull request describing **what** changed and **why**.

## Code style

- Idiomatic Go; run `gofmt` / `goimports`.
- Match the style of the surrounding code.
- New plugins: follow the patterns in `core/plugin/` and document them in
  [docs/PLUGINS.md](docs/PLUGINS.md).

## Reporting security issues

Do **not** open a public issue for security vulnerabilities. See
[SECURITY.md](SECURITY.md) for the private disclosure process.

## License

Contributions are made under the project license (GPLv3, see [LICENSE](LICENSE))
and the terms of the [CLA](CLA.md).
