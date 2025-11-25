#!/bin/bash
# Backup game data script for Matrix MUD

BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"

echo "💾 Creating backup in $BACKUP_DIR..."

mkdir -p "$BACKUP_DIR"
cp -r data/* "$BACKUP_DIR/"

echo "✅ Backup complete: $BACKUP_DIR"
echo "📊 Backup contents:"
ls -lh "$BACKUP_DIR"
