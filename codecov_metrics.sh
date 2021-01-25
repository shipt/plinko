#!/usr/bin/env bash
EXIT_CODE=0

echo "" > coverage.txt

for directory in `go list ./...`; do
  go test -coverprofile=profile.out $directory

  if [ $? -eq 1 ]; then
    EXIT_CODE=1
  fi

  if [ -f profile.out ] && [ $EXIT_CODE != 1 ]; then
    cat profile.out >> coverage.txt
    rm profile.out
  fi
done

if [ $EXIT_CODE == 1 ]; then
  exit $EXIT_CODE
fi

if [ -n "$CODECOV_TOKEN" ]; then
  curl -s https://codecov.io/bash | bash
else
  echo "CODECOV_TOKEN not set"
  EXIT_CODE=1
fi

rm coverage.txt
exit $EXIT_CODE
