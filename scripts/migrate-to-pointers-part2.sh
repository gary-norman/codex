#!/bin/bash
# Migration script Part 2: Update template struct field types
# Fixes struct literal assignment errors

set -e

echo "════════════════════════════════════════════════════════════"
echo "  Pointer Migration Script Part 2 - Template Structs"
echo "════════════════════════════════════════════════════════════"
echo ""

# Find all struct definitions with Channel/Post slice fields
echo "🔍 Finding template struct definitions..."
echo ""

# List files that likely contain template structs
FILES_TO_CHECK=(
    "internal/http/handlers/channel-handlers.go"
    "internal/http/handlers/home-handlers.go"
    "internal/http/handlers/post-handlers.go"
)

echo "📝 Files to update:"
for file in "${FILES_TO_CHECK[@]}"; do
    if [ -f "$file" ]; then
        echo "  - $file"
    fi
done
echo ""

echo "🔧 Applying changes..."
echo ""

# Update channel-handlers.go ownedChannels/joinedChannels variable declarations
sed -i.bak2 's/ownedChannels := make(\[\]models.Channel, 0)/ownedChannels := make([]*models.Channel, 0)/' \
    internal/http/handlers/channel-handlers.go 2>/dev/null || true

sed -i.bak2 's/joinedChannels := make(\[\]models.Channel, 0)/joinedChannels := make([]*models.Channel, 0)/' \
    internal/http/handlers/channel-handlers.go 2>/dev/null || true

sed -i.bak2 's/ownedAndJoinedChannels := make(\[\]models.Channel, 0)/ownedAndJoinedChannels := make([]*models.Channel, 0)/' \
    internal/http/handlers/channel-handlers.go 2>/dev/null || true

# Fix UpdateTimeSince calls for thisChannel
sed -i.bak2 's/models.UpdateTimeSince(&thisChannel)/models.UpdateTimeSince(thisChannel)/' \
    internal/http/handlers/channel-handlers.go 2>/dev/null || true

echo "✅ Variable declarations updated"
echo ""

# Clean up backup files
rm -f internal/http/handlers/*.bak2

echo "📊 Remaining issues require manual struct field updates:"
echo ""
echo "Look for struct definitions like:"
echo "  type SomeData struct {"
echo "    AllChannels    []models.Channel   // Change to []*models.Channel"
echo "    Posts          []models.Post      // Change to []*models.Post"
echo "  }"
echo ""
echo "Grep for these patterns:"
echo "  grep -n 'AllChannels.*\[\]models.Channel' internal/http/handlers/*.go"
echo "  grep -n 'Posts.*\[\]models.Post' internal/http/handlers/*.go"
echo ""

# Try to find struct definitions
echo "🔎 Struct fields that might need updating:"
grep -n "AllChannels\|OwnedChannels\|JoinedChannels\|Posts.*\[\]models" internal/http/handlers/*.go 2>/dev/null | head -20 || echo "  (None found with simple grep)"
echo ""

echo "════════════════════════════════════════════════════════════"
echo "Next: Find and update struct field types manually, then rebuild"
echo "════════════════════════════════════════════════════════════"
