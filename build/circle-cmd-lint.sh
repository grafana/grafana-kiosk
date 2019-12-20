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

go get -u github.com/jgautheron/goconst/cmd/goconst
go get -u honnef.co/go/tools/cmd/staticcheck
go get -u github.com/mgechev/revive
go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# use golangci-when possible
exit_if_fail golangci-lint run --deadline 10m --disable-all \
  --enable=deadcode\
  --enable=gofmt\
  --enable=ineffassign\
  --enable=structcheck\
  --enable=unconvert\
  --enable=varcheck

exit_if_fail go vet ./pkg/...

exit_if_fail revive -formatter stylish -config ./build/revive.toml
