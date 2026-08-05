#!/usr/bin/env bash
set -euo pipefail

# Build static GOPROXY files under local/goproxy.
# Usage:
#   ./scripts/build_goproxy.sh [vX.Y.Z]

# VERSION, GOPROXY_VERSION, and GOPROXY_MODULE_PATH may also be provided through
# the environment. Without an explicit version, an exact semver tag at HEAD is
# used; otherwise a Go pseudo-version is derived from the current commit.

REPO_DIR="$(git rev-parse --show-toplevel)"
cd "$REPO_DIR"

VERSION="${1:-${GOPROXY_VERSION:-${VERSION:-}}}"
OUT_DIR="${GOPROXY_OUT_DIR:-local/goproxy}"
MODULE_PATH="${GOPROXY_MODULE_PATH:-$(awk '/^module /{print $2}' go.mod)}"

if [[ -z "$MODULE_PATH" ]]; then
  echo "Could not read module path from go.mod"
  exit 1
fi

if [[ -z "$VERSION" ]]; then
  while IFS= read -r tag; do
    if [[ "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
      VERSION="$tag"
    fi
  done < <(git tag --points-at HEAD --sort=v:refname)
fi

if [[ -z "$VERSION" ]]; then
  COMMIT_TIME="$(TZ=UTC git show -s --format=%cd --date=format:%Y%m%d%H%M%S HEAD)"
  COMMIT_SHA="$(git rev-parse --short=12 HEAD)"
  VERSION="v0.0.0-${COMMIT_TIME}-${COMMIT_SHA}"
fi

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  echo "Invalid Go module version: $VERSION"
  echo "Expected a v-prefixed semver tag or pseudo-version."
  exit 1
fi

COMMIT_RFC3339="$(TZ=UTC git show -s --format=%cd --date=format:%Y-%m-%dT%H:%M:%SZ HEAD)"
VERSION_DIR="$OUT_DIR/$MODULE_PATH/@v"

rm -rf "$OUT_DIR"
mkdir -p "$VERSION_DIR"

cp go.mod "$VERSION_DIR/$VERSION.mod"
printf '{"Version":"%s","Time":"%s"}\n' "$VERSION" "$COMMIT_RFC3339" > "$VERSION_DIR/$VERSION.info"
git archive --format=zip --prefix="${MODULE_PATH}@${VERSION}/" HEAD > "$VERSION_DIR/$VERSION.zip"

{
  git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=v:refname
  printf '%s\n' "$VERSION"
} | awk '!seen[$0]++' > "$VERSION_DIR/list"

echo "Built GOPROXY files for ${MODULE_PATH}@${VERSION} at ${VERSION_DIR}"
