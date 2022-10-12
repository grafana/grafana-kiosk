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

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.0

go get github.com/jgautheron/goconst/cmd/goconst
go get honnef.co/go/tools/cmd/staticcheck
go get github.com/mgechev/revive
go install github.com/jgautheron/goconst/cmd/goconst
go install honnef.co/go/tools/cmd/staticcheck
go install github.com/mgechev/revive

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
  --enable=asciicheck\
  --enable=bodyclose\
  --enable=containedctx\
  --enable=contextcheck\
  --enable=decorder\
  --enable=depguard\
  --enable=dogsled\
  --enable=dupl\
  --enable=dupword\
  --enable=durationcheck\
  --enable=errchkjson\
  --enable=errname\
  --enable=errorlint\
  --enable=execinquery\
  --enable=exhaustive\
  --enable=exhaustruct\
  --enable=exportloopref\
  --enable=forbidigo\
  --enable=forcetypeassert\
  --enable=gochecknoglobals\
  --enable=gochecknoinits\
  --enable=gocognit\
  --enable=goconst\
  --enable=gocritic\
  --enable=gocyclo\
  --enable=gocognit\
  --enable=godot\
  --enable=godox\
  --enable=goerr113\
  --enable=gofmt\
  --enable=goheader\
  --enable=goprintffuncname\
  --enable=gosec\
  --enable=grouper\
  --enable=importas\
  --enable=interfacebloat\
  --enable=ireturn\
  --enable=loggercheck\
  --enable=maintidx\
  --enable=makezero\
  --enable=misspell\
  --enable=nakedret\
  --enable=nestif\
  --enable=nilerr\
  --enable=nilnil\
  --enable=nlreturn\
  --enable=noctx\
  --enable=nolintlint\
  --enable=nonamedreturns\
  --enable=nosprintfhostport\
  --enable=prealloc\
  --enable=predeclared\
  --enable=promlinter\
  --enable=reassign\
  --enable=revive\
  --enable=stylecheck\
  --enable=tenv\
  --enable=unconvert\
  --enable=usestdlibvars\
  --enable=varnamelen\
  --enable=whitespace\
  --enable=wrapcheck\
  --enable=wsl\
    ./...



# TODO: Enable these linters in the future
# --enable=cyclop\
# --enable=funlen
# --enable=gomnd\
# --enable=lll
# --enable=unparam
# --enable=paralleltest\
# --enable=tagliatelle\
# --enable=wsl

# go vet is already run by linter above
#exit_if_fail go vet ./pkg/...

exit_if_fail revive -formatter stylish -config ./build/revive.toml
