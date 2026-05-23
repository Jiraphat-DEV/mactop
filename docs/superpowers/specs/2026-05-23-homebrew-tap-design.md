# Design: Homebrew Tap for the Jiraphat-DEV mactop fork

**Date:** 2026-05-23
**Status:** Approved

## Goal

Let users install the customized `mactop` fork (`github.com/Jiraphat-DEV/mactop`,
with Plan A's Smart Alert Engine + `mactop status` and Plan B's background service
mode) through Homebrew, and document it in the README — without disturbing the
upstream `metaspartan/mactop` install path.

## Decisions (locked)

| Question | Decision | Rationale |
|----------|----------|-----------|
| Formula delivery | **Build from source** (`depends_on "go" => :build`) | Fork is personal; tagging is enough. No release publishing, no CGO/code-signing pipeline. Xcode CLT on the user's Mac supplies the toolchain. |
| Tap location | **New repo `Jiraphat-DEV/homebrew-mactop`** | Standard Homebrew tap layout (`homebrew-<name>`), clean `brew tap` + `brew install`. |
| Formula / binary name | **`mactop`** + `conflicts_with "mactop"` | Installed via the fully-qualified `jiraphat-dev/mactop/mactop`; the conflict guard prevents clobbering the homebrew-core `mactop` binary. |
| Version | **`v2.1.3-jiraphat.1`** | Keeps upstream `2.1.3` base, fork suffix makes it unambiguous that this is a fork ahead of upstream. |
| License | **MIT** | Matches `LICENSE` (MIT, © Carsen Klock). |

## Components

### 1. Version bump (this repo)
- `internal/app/globals.go:28` — `version` string → `"v2.1.3-jiraphat.1"`.
- Commit on `feat/homebrew-tap`, push, then annotated git tag `v2.1.3-jiraphat.1`, push tag.
- GitHub auto-generates the source tarball at
  `https://github.com/Jiraphat-DEV/mactop/archive/refs/tags/v2.1.3-jiraphat.1.tar.gz`.

### 2. Formula `mactop.rb` (in the tap repo)
Hand-written build-from-source formula (drops the GoReleaser "DO NOT EDIT" header):

```ruby
class Mactop < Formula
  desc "Apple Silicon Monitor Top (Jiraphat-DEV fork: alerts + background service)"
  homepage "https://github.com/Jiraphat-DEV/mactop"
  url "https://github.com/Jiraphat-DEV/mactop/archive/refs/tags/v2.1.3-jiraphat.1.tar.gz"
  sha256 "<computed from the real tarball after the tag is pushed>"
  license "MIT"

  depends_on "go" => :build
  depends_on :macos
  depends_on arch: :arm64

  conflicts_with "mactop", because: "both install a `mactop` binary"

  def install
    ENV["CGO_ENABLED"] = "1"
    system "go", "build", *std_go_args(output: bin/"mactop"), "main.go"
  end

  # service do / caveats: carried over from upstream, homepage updated to the fork.

  test do
    assert_match "v2.1.3-jiraphat.1", shell_output("#{bin}/mactop --version")
  end
end
```

Notes:
- No `go generate` step — the repo has zero `//go:generate` directives, so the
  GoReleaser `go generate ./...` before-hook was a no-op. The build is plain `go build`.
- `CGO_ENABLED=1` is required (Objective-C sources: `displayfps.m`, `ioreport.m`, `smc.h`).
- `service` / `caveats` blocks copied from the existing `mactop.rb`, homepage swapped to the fork.

### 3. Tap repo `Jiraphat-DEV/homebrew-mactop`
- Created with `gh repo create Jiraphat-DEV/homebrew-mactop --public` (asks for
  confirmation before the repo is actually created — outward-facing action).
- Contains `mactop.rb` + a short `README.md` (what the fork adds, the install command).
- Install path:
  ```bash
  brew tap jiraphat-dev/mactop
  brew install jiraphat-dev/mactop/mactop
  ```

### 4. README (this repo)
- Add a new section **"Install the Jiraphat-DEV fork (customized version)"** under the
  existing Homebrew section.
- List what the fork adds over upstream (Smart Alert Engine, `mactop status`,
  background service via `mactop install`/`uninstall`) plus the `brew tap`/`install` commands.
- The existing upstream Homebrew section is **kept**, not removed.

## Execution sequence

1. Bump version in `globals.go`, commit, tag `v2.1.3-jiraphat.1`, push branch + tag.
2. `curl -L` the GitHub tarball → compute `sha256`.
3. Write `mactop.rb` with the real sha256.
4. Create the tap repo (with confirmation) + push the formula.
5. `brew install jiraphat-dev/mactop/mactop` to verify it builds and installs.
6. Add the fork install section to this repo's README.

## Out of scope
- Prebuilt-binary releases / GoReleaser pipeline (deferred; revisit only if compile
  time on install becomes a complaint).
- Translating new README copy into the 19 i18n locales (README is English-only).
- Submitting the formula to homebrew-core.
