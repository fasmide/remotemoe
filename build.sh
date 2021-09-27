#!/bin/bash

if [ -z "$(git status --porcelain)" ]; then 
  # Uncommitted changes
  echo "build.sh will not build with uncommitted changes..."
  exit 1
fi

GITBRANCH=$(git branch --show-current)
GITHASH=$(git rev-list -1 HEAD)
GITDATE=$(git show -s --format=%ci ${GITHASH})
GITREPOSITORY=$(git config --get remote.origin.url)

go build \
  -ldflags "-X main.buildvars.GitDate=$GITDATE" \
  -ldflags "-X main.buildvars.GitHash=$GITHASH" \
  -ldflags "-X main.buildvars.GitBranch=$GITBRANCH" \
  -ldflags "-X main.buildvars.GitRepository=$GITREPOSITORY" \
  .