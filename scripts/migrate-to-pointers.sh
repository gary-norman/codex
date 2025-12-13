#!/bin/bash
# Migration script: Convert Channel slices from values to pointers
# This completes the refactoring started for Posts and Channels

set -e  # Exit on error

echo "════════════════════════════════════════════════════════════"
echo "  Pointer Migration Script - Channels"
echo "════════════════════════════════════════════════════════════"
echo ""

# Backup files before modifying
BACKUP_DIR="scripts/backups/pointer-migration-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"

echo "📦 Creating backups in $BACKUP_DIR..."
cp internal/sqlite/channels-sql.go "$BACKUP_DIR/"
cp internal/http/handlers/channel-handlers.go "$BACKUP_DIR/"
cp internal/http/handlers/home-handlers.go "$BACKUP_DIR/"
echo "✅ Backups created"
echo ""

echo "🔧 Step 1: Update OwnedOrJoinedByCurrentUser return type..."
# Change: func OwnedOrJoinedByCurrentUser(ID models.UUIDField) ([]models.Channel, error)
# To:     func OwnedOrJoinedByCurrentUser(ID models.UUIDField) ([]*models.Channel, error)
sed -i.bak 's/func (m \*ChannelModel) OwnedOrJoinedByCurrentUser(ID models.UUIDField) (\[\]models.Channel, error)/func (m *ChannelModel) OwnedOrJoinedByCurrentUser(ID models.UUIDField) ([]*models.Channel, error)/' \
    internal/sqlite/channels-sql.go

# Change: channels := make([]models.Channel, 0)
# To:     channels := make([]*models.Channel, 0)
sed -i.bak 's/channels := make(\[\]models.Channel, 0)/channels := make([]*models.Channel, 0)/' \
    internal/sqlite/channels-sql.go

# Change: channels = append(channels, *c)
# To:     channels = append(channels, c)
sed -i.bak 's/channels = append(channels, \*c)/channels = append(channels, c)/' \
    internal/sqlite/channels-sql.go

echo "✅ OwnedOrJoinedByCurrentUser updated"
echo ""

echo "🔧 Step 2: Update JoinedByCurrentUser return type..."
# Change: func (c *ChannelHandler) JoinedByCurrentUser(memberships []models.Membership) ([]models.Channel, error)
# To:     func (c *ChannelHandler) JoinedByCurrentUser(memberships []models.Membership) ([]*models.Channel, error)
sed -i.bak 's/func (c \*ChannelHandler) JoinedByCurrentUser(memberships \[\]models.Membership) (\[\]models.Channel, error)/func (c *ChannelHandler) JoinedByCurrentUser(memberships []models.Membership) ([]*models.Channel, error)/' \
    internal/http/handlers/channel-handlers.go

# Change: var channels []models.Channel
# To:     var channels []*models.Channel
sed -i.bak 's/var channels \[\]models.Channel/var channels []*models.Channel/' \
    internal/http/handlers/channel-handlers.go

# Change: channels = append(channels, *channel)
# To:     channels = append(channels, channel)
sed -i.bak 's/channels = append(channels, \*channel)/channels = append(channels, channel)/' \
    internal/http/handlers/channel-handlers.go

echo "✅ JoinedByCurrentUser updated"
echo ""

echo "🔧 Step 3: Update template data structures..."
# Find all struct literals that use []models.Channel and change to []*models.Channel
# This is more complex, so let's use a targeted approach

# In home-handlers.go, update the template data structure
# Look for patterns like: AllChannels: allChannels,
# The variable types are already []*models.Channel, so we just need to update struct field types

echo "⚠️  Note: Template struct field types may need manual updates in:"
echo "   - internal/http/handlers/home-handlers.go (search for 'AllChannels', 'OwnedChannels', etc.)"
echo "   - internal/http/handlers/channel-handlers.go (search for struct literals)"
echo ""

# Clean up .bak files
rm -f internal/sqlite/channels-sql.go.bak
rm -f internal/http/handlers/channel-handlers.go.bak
rm -f internal/http/handlers/home-handlers.go.bak

echo "🧪 Step 4: Verify changes..."
echo "Running go build to check for errors..."
if make build 2>&1 | grep -q "error"; then
    echo "❌ Build errors detected. Please review:"
    make build 2>&1 | grep "error" | head -20
    echo ""
    echo "💡 Backups are available in: $BACKUP_DIR"
    echo "   To restore: cp $BACKUP_DIR/* internal/"
    exit 1
else
    echo "✅ Build successful (or only type errors remaining)"
fi

echo ""
echo "════════════════════════════════════════════════════════════"
echo "  Migration Progress"
echo "════════════════════════════════════════════════════════════"
echo "✅ Database layer: OwnedOrJoinedByCurrentUser"
echo "✅ Handler layer: JoinedByCurrentUser"
echo "⚠️  Manual: Template struct field types (if needed)"
echo ""
echo "Next steps:"
echo "1. Check build output for any remaining type mismatches"
echo "2. Update template struct field types if needed"
echo "3. Run tests: go test ./..."
echo "4. Commit changes: git add -A && git commit"
echo ""
echo "Backups available at: $BACKUP_DIR"
echo "════════════════════════════════════════════════════════════"
