#!/bin/bash
find . -name "*.go" | sort | while read file; do
  echo "## $(basename "$file")" >> all_code.md
  echo '```go' >> all_code.md
  cat "$file" >> all_code.md
  echo '```' >> all_code.md
  echo "" >> all_code.md
done

