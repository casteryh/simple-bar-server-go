package main

var realms = []string{"yabai", "skhd", "aerospace", "widget", "missive"}

var yabaiKinds = []string{"spaces", "windows", "displays"}
var yabaiActions = []string{"refresh"}

var skhdKinds = []string{"mode"}
var skhdActions = []string{"refresh"}

var aerospaceKinds = []string{"spaces"}
var aerospaceActions = []string{"refresh"}

var widgetKinds = []string{
	"battery",
	"browser-track",
	"cpu",
	"crypto",
	"date-display",
	"keyboard",
	"github",
	"gpu",
	"memory",
	"mic",
	"mpd",
	"music",
	"netstats",
	"next-meeting",
	"notifications",
	"sound",
	"spotify",
	"stock",
	"time",
	"user-widget",
	"viscosity-vpn",
	"weather",
	"wifi",
	"youtube-music",
	"zoom",
}
var widgetActions = []string{"toggle", "enable", "disable", "refresh"}

var missiveActions = []string{"push"}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
