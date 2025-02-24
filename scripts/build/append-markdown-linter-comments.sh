#!/usr/bin/env bash

# Add markdown-linter comments to disable line-length and reference-links-images checks
sed -i "1i\<!-- markdownlint-disable line-length -->\n<!-- markdownlint-disable reference-links-images -->" docs/api.md
