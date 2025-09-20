#!/bin/bash

find . -type f -name "*.go" | while read -r file; do
    sed -i -E '/^[[:space:]]*\/\/.*\.\.\.$/d' "$file"
    sed -i -E ':a;N;$!ba;s/\n{3,}/\n\n/g' "$file"
done
