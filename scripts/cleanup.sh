#!/bin/bash
# Deep Tree Echo - Codebase Cleanup Script
# Moves duplicates and outdated files to archive, consolidates entry points

set -e

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ARCHIVE_DIR="$REPO_ROOT/archive"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "=========================================="
echo "Deep Tree Echo - Codebase Cleanup"
echo "=========================================="
echo "Repository: $REPO_ROOT"
echo "Timestamp: $TIMESTAMP"
echo ""

# Create archive subdirectories if needed
mkdir -p "$ARCHIVE_DIR/cleanup_$TIMESTAMP"
mkdir -p "$ARCHIVE_DIR/cleanup_$TIMESTAMP/cmd_duplicates"
mkdir -p "$ARCHIVE_DIR/cleanup_$TIMESTAMP/root_tests"
mkdir -p "$ARCHIVE_DIR/cleanup_$TIMESTAMP/server_duplicates"

# 1. Move duplicate main files from cmd/autonomous/
echo "[1/5] Archiving duplicate cmd/autonomous/ main files..."
for f in main_consolidated.go main_iteration001.go main_simple_consolidated.go; do
    if [ -f "$REPO_ROOT/cmd/autonomous/$f" ]; then
        mv "$REPO_ROOT/cmd/autonomous/$f" "$ARCHIVE_DIR/cleanup_$TIMESTAMP/cmd_duplicates/"
        echo "  Moved: cmd/autonomous/$f"
    fi
done

# 2. Move test files from root to archive
echo "[2/5] Archiving root-level test files..."
for f in "$REPO_ROOT"/test_*.go; do
    if [ -f "$f" ]; then
        mv "$f" "$ARCHIVE_DIR/cleanup_$TIMESTAMP/root_tests/"
        echo "  Moved: $(basename $f)"
    fi
done

# Move predict.go (broken, has build ignore)
if [ -f "$REPO_ROOT/predict.go" ]; then
    mv "$REPO_ROOT/predict.go" "$ARCHIVE_DIR/cleanup_$TIMESTAMP/"
    echo "  Moved: predict.go"
fi

# 3. Clean up old backup files (keep structure, just list)
echo "[3/5] Counting backup files in archive/backup_files/..."
BACKUP_COUNT=$(find "$ARCHIVE_DIR/backup_files" -name "*.bak" 2>/dev/null | wc -l)
echo "  Found $BACKUP_COUNT .bak files (preserved for reference)"

# 4. Remove empty directories
echo "[4/5] Cleaning empty directories..."
find "$REPO_ROOT" -type d -empty -not -path "*/.git/*" -delete 2>/dev/null || true

# 5. Verify build still works
echo "[5/5] Verifying build..."
cd "$REPO_ROOT"
if GOTOOLCHAIN=local CGO_ENABLED=0 go build ./... 2>/dev/null; then
    echo "  Build: SUCCESS"
else
    echo "  Build: FAILED (check errors)"
fi

echo ""
echo "=========================================="
echo "Cleanup Complete"
echo "=========================================="
echo "Archived to: $ARCHIVE_DIR/cleanup_$TIMESTAMP/"
echo ""
echo "Next steps:"
echo "  1. Review archived files"
echo "  2. Run: docker-compose -f docker-compose.production.yml up -d"
echo "  3. Check health: curl http://localhost:8080/health"
