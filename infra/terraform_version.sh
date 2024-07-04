#!/bin/sh
set -eu
grep -oP '(?<=required_version = "~> )[0-9]+\.[0-9]+\.[0-9]+(?=")' main.tf
