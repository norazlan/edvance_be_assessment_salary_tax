#!/bin/bash

while getopts ":wl" opt; do
  case ${opt} in
    w )
      echo "Building for Windows"
      GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-cc go build -o ./dist/payslip.exe app.go
      cp .env.sample ./dist/.env
      exit 0
      ;;
    l )
      echo "Building for Linux"
      GOOS=linux GOARCH=amd64 go build -o ./dist/payslip app.go
      cp .env.sample ./dist/.env
      exit 0
      ;;
  esac
done

if [ $OPTIND -eq 1 ]; then
  echo "No options were passed"
  echo "Usage for windows build: $0 -w" >&2
  echo "Usage for linux build: $0 -l">&2
  exit 1
fi