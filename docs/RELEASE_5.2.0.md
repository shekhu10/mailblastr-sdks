# MailBlastr SDK release 5.2.0

Status: prepared and locally verified; publication is held for the backend prerequisite.

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

## Backend prerequisite

Production main was checked at `0d6f2203534a6122a835ac742b9196bc8b9c5bfc`.
GET /api/domains/sdk-release-probe/tracking-health returned an HTML 404, confirming
that the new tracking-health route is absent. The backend fixes, including reply
and forward operation keys, remain local in the webapp repository. The backend
release procedure requires migrations through 190, the matching worker, then web.
Authorization to extend this SDK release to those production changes is pending.

## Publishing

The repository Release workflow publishes Node, CLI, PyPI, RubyGems, crates.io,
NuGet, Maven Central, the PHP subtree/Packagist package, and the Go module tag.
The previous 5.1.1 run succeeded for all nine. Existing GitHub secrets and
trusted publishers are configured; credentials are not stored in this repository.

After the backend prerequisite is satisfied, push v5.2.0 and verify every registry
against that exact version. Record the release run and registry results here.
