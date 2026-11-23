#!/usr/bin/env bash
# Merge charmbracelet/crush main into this fork and push it.
set -euo pipefail
git remote get-url upstream >/dev/null 2>&1 || git remote add upstream https://github.com/charmbracelet/crush.git
git fetch upstream main
git checkout main
git pull --ff-only origin main
git merge upstream/main
go build ./...
git push origin main
