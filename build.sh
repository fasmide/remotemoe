#!/bin/bash -e

GITBRANCH=$(git branch --show-current)
GITHASH=$(git rev-list -1 HEAD)
GITDATE=$(git show -s --format=%ci ${GITHASH})
GITREPOSITORY=$(git config --get remote.origin.url)
GITPORCELAIN=$(git status --porcelain)

go build "$@" \
  -ldflags "-X github.com/fasmide/remotemoe/buildvars.Initialized=true
  -X github.com/fasmide/remotemoe/buildvars.GitCommit=${GITHASH}
  -X 'github.com/fasmide/remotemoe/buildvars.GitCommitDate=${GITDATE}' 
  -X github.com/fasmide/remotemoe/buildvars.GitBranch=${GITBRANCH}
  -X github.com/fasmide/remotemoe/buildvars.GitRepository=${GITREPOSITORY}
  -X 'github.com/fasmide/remotemoe/buildvars.GitPorcelain=${GITPORCELAIN}'" \
  .
