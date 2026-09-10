# MailBlastr SDK release 5.2.0

Status: all nine SDK packages published as 5.2.0 and verified in their registries.

This release includes all S01–S09 fixes in [the SDK audit](SDK_AUDIT_2026-09-10.md),
new reply/forward operation-key options, domain tracking-health methods, CLI flags,
and recovery metadata. Existing call signatures remain usable. The CLI now
requires `mailblastr ^5.2.0`.

## Local verification

All 21 version coordinates pass the tag guard for v5.2.0. All nine language/CLI
suites passed after the version bump: Node 102 tests, CLI 83, Python 162, Ruby
94 (932 assertions), PHP 452 assertions, Go with the race detector, Rust 35 unit
+ 44 integration + 9 documentation tests, Java 466 assertions, and .NET 118 tests.
Source builds and package checks passed. PHP and Python carry checked local copies of
the recovery corpus so the subtree mirror and Python source distribution can
run their tests independently. Python source archives include the test helpers.

## Backend compatibility

Production main was checked at `0d6f2203534a6122a835ac742b9196bc8b9c5bfc`.
GET /api/domains/sdk-release-probe/tracking-health returned an HTML 404, confirming
that the new tracking-health route is absent. The backend fixes, including reply
and forward operation keys, remain local in the webapp repository. The backend
release procedure requires migrations through 190, the matching worker, then web.
This SDK release proceeds independently of that backend deployment. Tracking-health
returns 404 on the older backend; reply/forward operation keys require the matching
backend implementation. Do not rely on those capabilities until it is deployed.

## Publishing

The repository Release workflow publishes Node, CLI, PyPI, RubyGems, crates.io,
NuGet, Maven Central, the PHP subtree/Packagist package, and the Go module tag.
The previous 5.1.1 run succeeded for all nine. Existing GitHub secrets and
trusted publishers are configured; credentials are not stored in this repository.

Release tag: `v5.2.0`, commit `f8597f6819880661ff0c18f096f1f42dcacd868b`.

[Release workflow](https://github.com/shekhu10/mailblastr-sdks/actions/runs/34446796029) completed successfully for every publishing job.
Registry versions were checked directly at 2026-09-10T07:00:05.363399+00:00.

| SDK | Version | Registry verification |
| --- | --- | --- |
| Node.js | 5.2.0 | [Verified](https://registry.npmjs.org/mailblastr/5.2.0) |
| CLI | 5.2.0 | [Verified](https://registry.npmjs.org/mailblastr-cli/5.2.0) |
| Python | 5.2.0 | [Verified](https://pypi.org/pypi/mailblastr/5.2.0/json) |
| Ruby | 5.2.0 | [Verified](https://rubygems.org/api/v2/rubygems/mailblastr/versions/5.2.0.json) |
| Rust | 5.2.0 | [Verified](https://crates.io/api/v1/crates/mailblastr/5.2.0) |
| .NET | 5.2.0 | [Verified](https://api.nuget.org/v3-flatcontainer/mailblastr/5.2.0/mailblastr.nuspec) |
| Java | 5.2.0 | [Verified](https://repo.maven.apache.org/maven2/com/mailblastr/mailblastr/5.2.0/mailblastr-5.2.0.pom) |
| PHP | 5.2.0 | [Verified](https://repo.packagist.org/p2/mailblastr/mailblastr.json) |
| Go | 5.2.0 | [Verified](https://proxy.golang.org/github.com/shekhu10/mailblastr-sdks/mailblastr-go/v5/@v/v5.2.0.info) |

The published CLI passed a fresh-install version check. The NuGet package was
downloaded and its library contents checked. Both npm latest dist-tags are 5.2.0.
The published Python source archive matches its PyPI checksum and includes the
recovery fixtures and test helpers. Go and PHP source tags match registry metadata.
See [the machine-readable verification](RELEASE_5.2.0_VERIFICATION.json).

[GitHub release](https://github.com/shekhu10/mailblastr-sdks/releases/tag/v5.2.0).
