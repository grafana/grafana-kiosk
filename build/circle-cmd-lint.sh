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

# gometalinter needs newer version of go-i18n
#go get -u github.com/nicksnyder/go-i18n/v2/i18n
go get -u github.com/nicksnyder/go-i18n
go get -u github.com/alecthomas/gometalinter
go get -u github.com/jgautheron/goconst/cmd/goconst
go get -u honnef.co/go/tools/cmd/staticcheck
go get -u github.com/mgechev/revive
go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

# use gometalinter when lints are not available in golangci or
# when gometalinter is better. Eg. goconst for gometalinter does not lint test files
# which is not desired.
exit_if_fail gometalinter --enable-gc --vendor --deadline 10m --disable-all \
  --enable=goconst\
  --enable=staticcheck

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
