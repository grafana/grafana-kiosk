#!/bin/bash

function exit_if_fail {
  command=$@
  echo "Executing '$command'"
  eval $command
  rc=$?
  if [ $rc -ne 0 ]; then
      echo "'$command' returned $rc."
      exit $rc
  fi
}

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0

go get -u github.com/jgautheron/goconst/cmd/goconst
go get -u honnef.co/go/tools/cmd/staticcheck
go get -u github.com/mgechev/revive
#go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# use golangci-when possible
# exit_if_fail golangci-lint run --deadline 10m --disable-all \
#  --enable=deadcode\
#  --enable=gofmt\
#  --enable=ineffassign\
#  --enable=structcheck\
#  --enable=unconvert\
#  --enable=varcheck

exit_if_fail golangci-lint --verbose run\
  --deadline 5m\
  --enable=bodyclose\
  --enable=gosec\
  --enable=interfacer\
  --enable=unconvert\
  --enable=dupl\
  --enable=goconst\
  --enable=gocyclo\
  --enable=gocognit\
  --enable=gofmt\
  --enable=maligned\
  --enable=depguard\
  --enable=misspell\
  --enable=dogsled\
  --enable=nakedret\
  --enable=prealloc\
  --enable=scopelint\
  --enable=gocritic\
  --enable=gochecknoinits\
  --enable=godox\
  --enable=whitespace\
    ./...

# TODO: Enable these linters in the future
# --enable=funlen
# --enable=gochecknoglobals
# --enable=lll
# --enable=unparam
# --enable=wsl

# go vet is already run by linter above
#exit_if_fail go vet ./pkg/...

exit_if_fail revive -formatter stylish -config ./build/revive.toml
