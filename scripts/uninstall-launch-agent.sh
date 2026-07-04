#!/bin/sh
set -eu

label="io.github.casteryh.simple-bar-server-go"
plist="$HOME/Library/LaunchAgents/$label.plist"

launchctl bootout "gui/$(id -u)" "$plist" 2>/dev/null || true
rm -f "$plist"

echo "Uninstalled $label"
