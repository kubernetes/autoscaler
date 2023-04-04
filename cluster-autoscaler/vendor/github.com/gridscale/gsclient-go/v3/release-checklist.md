# Release Checklist

## Prepare the release

- [ ] Make sure to check direct dependencies and keep the vendor directory current. Run `go mod tidy` and `go mod vendor` to make sure.
- [ ] Make sure test are passing. Run `go test $(go list ./... | grep -v /examples)`.
- [ ] Make sure the `version` number in `config.go` is correct.
- [ ] Make sure changelog is up-to-date and has a release date.
- [ ] Add a git tag for the release, e.g., `git tag v3.3.0`.
- [ ] Finally, push the tag, e.g., `git push origin v3.3.0`.

## Release and verify

- [ ] Finally, go to [GitHub](https://github.com/gridscale/gsclient-go/releases/) and finish up the draft release. Include the changelog entries in the release message.
