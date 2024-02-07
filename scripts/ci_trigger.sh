#!/bin/sh
set -e

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )";

NOW=$(date +%Y.%-m%d.%-H%M)

git commit --allow-empty -m "ci($NOW): ✨🐛🚨"

TARGET=${1:-origin}
echo "---------------------------"
printf "Pushing... $NOW --> %s\n" "$TARGET"
echo "---------------------------"
git push "$TARGET"  