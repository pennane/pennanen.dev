#!/bin/sh
since="${VERCEL_GIT_PREVIOUS_SHA:-HEAD^}"
git cat-file -e "$since^{commit}" 2>/dev/null || exit 1
git diff --quiet "$since" HEAD -- . ../../package.json ../../pnpm-lock.yaml ../../pnpm-workspace.yaml
