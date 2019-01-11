#!/bin/sh
#
# Runs shellcheck on all shell scripts in the istio repository.

TOOLS_DIR="$(cd "$(dirname "${0}")" && pwd -P)"
ISTIO_ROOT="$(cd "$(dirname "${TOOLS_DIR}")" && pwd -P)"

# See https://github.com/koalaman/shellcheck/wiki for details on each code's
# corresponding rule.
EXCLUDES="1054,"
EXCLUDES="${EXCLUDES}1056,"
EXCLUDES="${EXCLUDES}1083,"
EXCLUDES="${EXCLUDES}1090,"
EXCLUDES="${EXCLUDES}1091,"
EXCLUDES="${EXCLUDES}1117,"
EXCLUDES="${EXCLUDES}2001,"
EXCLUDES="${EXCLUDES}2004,"
EXCLUDES="${EXCLUDES}2009,"
EXCLUDES="${EXCLUDES}2016,"
EXCLUDES="${EXCLUDES}2035,"
EXCLUDES="${EXCLUDES}2046,"
EXCLUDES="${EXCLUDES}2048,"
EXCLUDES="${EXCLUDES}2059,"
EXCLUDES="${EXCLUDES}2068,"
EXCLUDES="${EXCLUDES}2086,"
EXCLUDES="${EXCLUDES}2097,"
EXCLUDES="${EXCLUDES}2098,"
EXCLUDES="${EXCLUDES}2100,"
EXCLUDES="${EXCLUDES}2124,"
EXCLUDES="${EXCLUDES}2129,"
EXCLUDES="${EXCLUDES}2145,"
EXCLUDES="${EXCLUDES}2148,"
EXCLUDES="${EXCLUDES}2155,"
EXCLUDES="${EXCLUDES}2162,"
EXCLUDES="${EXCLUDES}2164,"
EXCLUDES="${EXCLUDES}2166,"
EXCLUDES="${EXCLUDES}2191,"
EXCLUDES="${EXCLUDES}2206,"
EXCLUDES="${EXCLUDES}2230,"
EXCLUDES="${EXCLUDES}2231"

# All files ending in .sh.
SH_FILES=$( \
    find "${ISTIO_ROOT}" \
        -name '*.sh' -type f \
        -not -path '*/vendor/*' \
        -not -path '*/.git/*')
# All files not ending in .sh but starting with a shebang.
SHEBANG_FILES=$( \
    find "${ISTIO_ROOT}" \
        -not -name '*.sh' -type f \
        -not -path '*/vendor/*' \
        -not -path '*/.git/*' | \
        while read f; do
            head -n 1 "$f" | grep -q '^#!.*sh' && echo "$f";
        done)

echo "${SH_FILES}" "${SHEBANG_FILES}" \
    | xargs shellcheck --exclude="${EXCLUDES}"
