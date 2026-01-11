#!/bin/bash
set -e

# Generate release notes from git commits using conventional commit format
# Usage: ./generate-release-notes.sh <version> <previous-tag>

VERSION="${1:-}"
PREV_TAG="${2:-}"
REPO="${GITHUB_REPOSITORY:-fedanant/asyncapi-doc}"

if [ -z "$VERSION" ]; then
  echo "Error: VERSION is required"
  echo "Usage: $0 <version> [previous-tag]"
  exit 1
fi

# Get previous tag if not provided
if [ -z "$PREV_TAG" ]; then
  PREV_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
fi

echo "Generating release notes for $VERSION"
if [ -n "$PREV_TAG" ]; then
  echo "Comparing against: $PREV_TAG"
else
  echo "No previous tag found - generating initial release notes"
fi

# Initialize output file
OUTPUT_FILE="release_notes.md"
> "$OUTPUT_FILE"

# Header
if [ -z "$PREV_TAG" ]; then
  echo "## 🎉 Initial Release" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  echo "This is the first release of asyncapi-doc, a CLI tool for generating AsyncAPI documentation from Go code." >> "$OUTPUT_FILE"
else
  echo "## What's Changed" >> "$OUTPUT_FILE"
fi

echo "" >> "$OUTPUT_FILE"

# Parse commits and categorize them
declare -A FEATURES=()
declare -A FIXES=()
declare -A BREAKING=()
declare -A DOCS=()
declare -A CHORES=()
declare -A OTHER=()

# Read commits
if [ -n "$PREV_TAG" ]; then
  COMMIT_RANGE="${PREV_TAG}..HEAD"
else
  COMMIT_RANGE="HEAD"
fi

while IFS= read -r line; do
  COMMIT_HASH=$(echo "$line" | cut -d'|' -f1)
  COMMIT_MSG=$(echo "$line" | cut -d'|' -f2-)

  # Parse conventional commit format
  if [[ "$COMMIT_MSG" =~ ^feat(\(.*\))?!?:\ (.+)$ ]]; then
    FEATURES["$COMMIT_HASH"]="$COMMIT_MSG"
    if [[ "$COMMIT_MSG" =~ ! ]]; then
      BREAKING["$COMMIT_HASH"]="$COMMIT_MSG"
    fi
  elif [[ "$COMMIT_MSG" =~ ^fix(\(.*\))?:\ (.+)$ ]]; then
    FIXES["$COMMIT_HASH"]="$COMMIT_MSG"
  elif [[ "$COMMIT_MSG" =~ ^docs(\(.*\))?:\ (.+)$ ]]; then
    DOCS["$COMMIT_HASH"]="$COMMIT_MSG"
  elif [[ "$COMMIT_MSG" =~ ^chore(\(.*\))?:\ (.+)$ ]]; then
    CHORES["$COMMIT_HASH"]="$COMMIT_MSG"
  elif [[ "$COMMIT_MSG" =~ ^perf(\(.*\))?:\ (.+)$ ]]; then
    FEATURES["$COMMIT_HASH"]="$COMMIT_MSG"
  elif [[ "$COMMIT_MSG" =~ ^refactor(\(.*\))?:\ (.+)$ ]]; then
    CHORES["$COMMIT_HASH"]="$COMMIT_MSG"
  elif [[ "$COMMIT_MSG" =~ ^test(\(.*\))?:\ (.+)$ ]]; then
    CHORES["$COMMIT_HASH"]="$COMMIT_MSG"
  elif [[ "$COMMIT_MSG" =~ BREAKING\ CHANGE ]]; then
    BREAKING["$COMMIT_HASH"]="$COMMIT_MSG"
    OTHER["$COMMIT_HASH"]="$COMMIT_MSG"
  else
    OTHER["$COMMIT_HASH"]="$COMMIT_MSG"
  fi
done < <(git log --pretty=format:"%h|%s" "$COMMIT_RANGE")

# Generate categorized changelog
if [ ${#BREAKING[@]} -gt 0 ]; then
  echo "### ⚠️ Breaking Changes" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  for hash in "${!BREAKING[@]}"; do
    msg="${BREAKING[$hash]}"
    echo "- $msg (\`$hash\`)" >> "$OUTPUT_FILE"
  done
  echo "" >> "$OUTPUT_FILE"
fi

if [ ${#FEATURES[@]} -gt 0 ]; then
  echo "### ✨ Features" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  for hash in "${!FEATURES[@]}"; do
    msg="${FEATURES[$hash]}"
    # Clean up the message - remove type prefix
    clean_msg=$(echo "$msg" | sed -E 's/^(feat|perf)(\([^)]+\))?!?: //')
    echo "- $clean_msg (\`$hash\`)" >> "$OUTPUT_FILE"
  done
  echo "" >> "$OUTPUT_FILE"
fi

if [ ${#FIXES[@]} -gt 0 ]; then
  echo "### 🐛 Bug Fixes" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  for hash in "${!FIXES[@]}"; do
    msg="${FIXES[$hash]}"
    # Clean up the message - remove type prefix
    clean_msg=$(echo "$msg" | sed -E 's/^fix(\([^)]+\))?: //')
    echo "- $clean_msg (\`$hash\`)" >> "$OUTPUT_FILE"
  done
  echo "" >> "$OUTPUT_FILE"
fi

if [ ${#DOCS[@]} -gt 0 ]; then
  echo "### 📚 Documentation" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  for hash in "${!DOCS[@]}"; do
    msg="${DOCS[$hash]}"
    clean_msg=$(echo "$msg" | sed -E 's/^docs(\([^)]+\))?: //')
    echo "- $clean_msg (\`$hash\`)" >> "$OUTPUT_FILE"
  done
  echo "" >> "$OUTPUT_FILE"
fi

if [ ${#OTHER[@]} -gt 0 ]; then
  echo "### 🔧 Other Changes" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
  for hash in "${!OTHER[@]}"; do
    msg="${OTHER[$hash]}"
    echo "- $msg (\`$hash\`)" >> "$OUTPUT_FILE"
  done
  echo "" >> "$OUTPUT_FILE"
fi

# Contributors section
echo "### 👥 Contributors" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
if [ -n "$PREV_TAG" ]; then
  git log --pretty=format:"%an" "$COMMIT_RANGE" | sort -u | while read -r contributor; do
    echo "- $contributor" >> "$OUTPUT_FILE"
  done
else
  git log --pretty=format:"%an" HEAD | sort -u | while read -r contributor; do
    echo "- $contributor" >> "$OUTPUT_FILE"
  done
fi
echo "" >> "$OUTPUT_FILE"

# Full changelog link
if [ -n "$PREV_TAG" ]; then
  echo "**Full Changelog**: https://github.com/$REPO/compare/${PREV_TAG}...${VERSION}" >> "$OUTPUT_FILE"
  echo "" >> "$OUTPUT_FILE"
fi

echo "---" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Installation section
cat >> "$OUTPUT_FILE" << 'EOF'
## 📦 Downloads

### Linux
- [asyncapi-doc-${VERSION}-linux-amd64.tar.gz](https://github.com/${REPO}/releases/download/${VERSION}/asyncapi-doc-${VERSION}-linux-amd64.tar.gz)
- [asyncapi-doc-${VERSION}-linux-arm64.tar.gz](https://github.com/${REPO}/releases/download/${VERSION}/asyncapi-doc-${VERSION}-linux-arm64.tar.gz)

### macOS
- [asyncapi-doc-${VERSION}-darwin-amd64.tar.gz](https://github.com/${REPO}/releases/download/${VERSION}/asyncapi-doc-${VERSION}-darwin-amd64.tar.gz)
- [asyncapi-doc-${VERSION}-darwin-arm64.tar.gz](https://github.com/${REPO}/releases/download/${VERSION}/asyncapi-doc-${VERSION}-darwin-arm64.tar.gz)

### Windows
- [asyncapi-doc-${VERSION}-windows-amd64.zip](https://github.com/${REPO}/releases/download/${VERSION}/asyncapi-doc-${VERSION}-windows-amd64.zip)
- [asyncapi-doc-${VERSION}-windows-arm64.zip](https://github.com/${REPO}/releases/download/${VERSION}/asyncapi-doc-${VERSION}-windows-arm64.zip)

### Checksums
SHA256 checksums are available for all downloads (*.sha256 files).

## 🚀 Installation

### Quick Install (Linux/macOS)
```bash
# Download and install the latest version
curl -fsSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash
```

### Manual Installation

#### Linux/macOS
```bash
# Download and extract
tar -xzf asyncapi-doc-${VERSION}-<os>-<arch>.tar.gz

# Make executable and move to PATH
chmod +x asyncapi-doc
sudo mv asyncapi-doc /usr/local/bin/
```

#### Windows
```powershell
# Extract the ZIP file
# Add the directory to your PATH or move asyncapi-doc.exe to a directory in your PATH
```

### Verify Installation
```bash
asyncapi-doc version
```

## 📖 Usage

```bash
# Generate AsyncAPI specification from Go code
asyncapi-doc generate -output ./asyncapi.yaml ./path/to/code

# Show help
asyncapi-doc --help
```
EOF

# Substitute variables in the template
sed -i.bak "s|\${VERSION}|$VERSION|g" "$OUTPUT_FILE"
sed -i.bak "s|\${REPO}|$REPO|g" "$OUTPUT_FILE"
rm -f "$OUTPUT_FILE.bak"

echo "✓ Release notes generated: $OUTPUT_FILE"
