#!/usr/bin/env bash

set -ex

rm -rf image/bingo/*
rm -rf image/test
rm -rf uploadsGO/*

docker rmi local/go_faas:0.0.1