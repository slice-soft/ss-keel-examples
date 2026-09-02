#!/usr/bin/env bash
#
# Bump every example to the latest published ss-keel-* release and prove the
# examples still build against it.
#
# Each example under examples/ is its own Go module. They drifted for months
# behind the released addons, so a reader copying from an example was copying an
# old API. This runs on a schedule and opens a PR when a release has moved.
#
# The build check is the point: a PR that bumps versions without compiling is
# worse than the drift it fixes.
#
# Requires: go. Run from the repository root.

set -euo pipefail

MODULE_PREFIX="github.com/slice-soft/ss-keel-"
EXAMPLES_DIR="examples"

bumped=0
failed=0
declare -a broken=()

for example_dir in "${EXAMPLES_DIR}"/*/; do
    example=$(basename "$example_dir")
    [[ -f "${example_dir}go.mod" ]] || continue

    # Every ss-keel-* module the example requires, direct or indirect.
    mapfile -t modules < <(grep -oE "${MODULE_PREFIX}[a-z]+" "${example_dir}go.mod" | sort -u)
    if [[ ${#modules[@]} -eq 0 ]]; then
        continue
    fi

    echo "── ${example}"
    (
        cd "$example_dir"
        for module in "${modules[@]}"; do
            if ! go get "${module}@latest"; then
                echo "ERROR ${example}: go get ${module}@latest failed" >&2
                exit 1
            fi
        done
        go mod tidy
    ) || { failed=$((failed + 1)); broken+=("$example (dependency update)"); continue; }

    # A bump that does not compile must not reach a PR unnoticed.
    if ! (cd "$example_dir" && go build ./... && go vet ./...); then
        echo "ERROR ${example}: does not build against the new versions" >&2
        failed=$((failed + 1))
        broken+=("$example (build)")
        continue
    fi

    # `go build` leaves a binary named after the module in the example directory.
    # Remove exactly that file: `git clean -fX` here would also delete a local
    # .env someone copied from .env.example to run the example.
    binary=$(awk '/^module /{print $2}' "${example_dir}go.mod" | xargs -r basename)
    if [[ -n "$binary" && -f "${example_dir}${binary}" ]]; then
        rm -f "${example_dir}${binary}"
    fi

    versions=$(grep -oE "ss-keel-[a-z]+ v[0-9]+\.[0-9]+\.[0-9]+" "${example_dir}go.mod" | tr '\n' ' ')
    echo "   ok  ${versions}"
    bumped=$((bumped + 1))
done

echo
echo "${bumped} example(s) verified, ${failed} failing."

if [[ ${failed} -gt 0 ]]; then
    printf 'Broken: %s\n' "${broken[*]}" >&2
    exit 1
fi
