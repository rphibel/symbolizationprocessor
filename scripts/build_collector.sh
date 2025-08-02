#!/bin/bash

set -e
set -x

TEST_DIR=$(mktemp -d)
cp builder-config.yaml "$TEST_DIR"
mkdir "$TEST_DIR/symbolizationprocessor"
cp -r * "$TEST_DIR/symbolizationprocessor/"
cp -r testdata/otelcol-dev "$TEST_DIR/otelcol-dev"

cd "$TEST_DIR"

go work init
go work use otelcol-dev
go work use symbolizationprocessor

go build -v ./otelcol-dev