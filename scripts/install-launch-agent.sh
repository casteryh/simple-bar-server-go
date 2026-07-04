#!/bin/sh
set -eu

label="io.github.casteryh.simple-bar-server-go"
repo_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
binary="$repo_dir/simple-bar-server"
plist="$HOME/Library/LaunchAgents/$label.plist"
stdout_log="/tmp/simple-bar-server-go.out.log"
stderr_log="/tmp/simple-bar-server-go.err.log"

go build -o "$binary" "$repo_dir"
mkdir -p "$(dirname -- "$plist")"

xml_escape() {
  printf '%s' "$1" | sed \
    -e 's/&/\&amp;/g' \
    -e 's/</\&lt;/g' \
    -e 's/>/\&gt;/g'
}

binary_xml="$(xml_escape "$binary")"
repo_dir_xml="$(xml_escape "$repo_dir")"
stdout_xml="$(xml_escape "$stdout_log")"
stderr_xml="$(xml_escape "$stderr_log")"

cat >"$plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>$label</string>

    <key>ProgramArguments</key>
    <array>
      <string>$binary_xml</string>
    </array>

    <key>WorkingDirectory</key>
    <string>$repo_dir_xml</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

    <key>ThrottleInterval</key>
    <integer>5</integer>

    <key>StandardOutPath</key>
    <string>$stdout_xml</string>

    <key>StandardErrorPath</key>
    <string>$stderr_xml</string>
  </dict>
</plist>
EOF

launchctl bootout "gui/$(id -u)" "$plist" 2>/dev/null || true
launchctl bootstrap "gui/$(id -u)" "$plist"
launchctl kickstart -k "gui/$(id -u)/$label"

echo "Installed $label"
