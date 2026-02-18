# huidi-led

[![Go Reference](https://pkg.go.dev/badge/github.com/alparslanahmed/huidi-led.svg)](https://pkg.go.dev/github.com/alparslanahmed/huidi-led)
[![Go Report Card](https://goreportcard.com/badge/github.com/alparslanahmed/huidi-led)](https://goreportcard.com/report/github.com/alparslanahmed/huidi-led)

A Go library for communicating with **Huidu full-color LED controller cards** via the Huidu SDK 2.0 binary TCP protocol.

## Features

- **Device Management** — Query device info (CPU, model, screen size, firmware version)
- **Display Programs** — Send text, image, video, and clock programs with 30 transition effects
- **Brightness Control** — Manual, scheduled, and sensor-based brightness management
- **Screen Control** — Open/close screen, scheduled on/off timers
- **Network Configuration** — Ethernet and WiFi settings (AP & Station modes)
- **Time Sync** — Synchronize device clock
- **File Transfer** — Upload images, videos, fonts, firmware with resume support and progress callbacks
- **File Management** — List and delete files on the device
- **Heartbeat** — Automatic connection keep-alive
- **Thread-Safe** — All public methods are mutex-protected; safe for concurrent use

## Installation

```bash
go get github.com/alparslanahmed/huidi-led
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	huidu "github.com/alparslanahmed/huidi-led"
)

func main() {
	// Create device and connect
	device := huidu.NewDevice("192.168.6.1", 10001)
	if err := device.Connect(); err != nil {
		log.Fatal(err)
	}
	defer device.Close()

	// Get device info
	info, err := device.GetDeviceInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Device: %s, Screen: %dx%d\n", info.DeviceID, info.ScreenWidth, info.ScreenHeight)

	// Send a text program
	screen := huidu.NewScreen()
	prog := screen.AddProgram("hello")
	area := prog.AddFullScreenArea(64, 32)
	area.AddText("Hello World!", huidu.TextConfig{
		Color:    "#ff0000",
		FontSize: 14,
		Effect:   huidu.EffectLeftScrollLoop,
		Speed:    4,
	})

	if err := device.SendScreen(screen); err != nil {
		log.Fatal(err)
	}
}
```

## Usage Examples

### Brightness Control

```go
// Set brightness to 80%
device.SetBrightness(80)

// Get current brightness info
info, _ := device.GetLuminanceInfo()
fmt.Printf("Mode: %d, Default: %d%%\n", info.Mode, info.DefaultValue)
```

### Screen On/Off

```go
device.OpenScreen()
device.CloseScreen()
```

### Scrolling Text

```go
screen := huidu.NewScreen()
prog := screen.AddProgram("scroll")
area := prog.AddFullScreenArea(64, 32)
area.AddText("Breaking News!", huidu.TextConfig{
	Color:    "#00ff00",
	FontSize: 14,
	Bold:     true,
	Effect:   huidu.EffectLeftScrollLoop,
	Speed:    3,
})
device.SendScreen(screen)
```

### Digital Clock

```go
screen := huidu.NewScreen()
prog := screen.AddProgram("clock")
area := prog.AddFullScreenArea(64, 32)
area.AddClock(huidu.ClockConfig{
	Type:       huidu.ClockDigital,
	ShowDate:   true,
	DateColor:  "#ffff00",
	DateFormat: 1,
	ShowTime:   true,
	TimeColor:  "#00ffff",
	TimeFormat: 1,
})
device.SendScreen(screen)
```

### File Upload with Progress

```go
device := huidu.NewDevice("192.168.6.1", 10001,
	huidu.WithProgressCallback(func(p huidu.UploadProgress) {
		fmt.Printf("\rUploading: %s [%.1f%%]", p.FileName, p.Percent)
	}),
)
device.Connect()
device.UploadFile("/path/to/image.jpg")
```

### Network Info

```go
eth, _ := device.GetEthernetInfo()
fmt.Printf("IP: %s, DHCP: %v\n", eth.IP, eth.AutoDHCP)

wifi, _ := device.GetWifiInfo()
fmt.Printf("WiFi: %v, AP SSID: %s\n", wifi.Enabled, wifi.APInfo.SSID)
```

### Time Sync

```go
device.SetTimeInfo(&huidu.TimeInfo{
	Time:     "2026-02-18 13:00:00",
	Timezone: "(UTC+03:00)Istanbul",
	Sync:     "none",
})
```

## Protocol Overview

The library implements the Huidu SDK 2.0 binary TCP protocol:

| Step | Description |
|------|-------------|
| 1 | TCP connection on port 10001 |
| 2 | Transport version negotiation (`0x2001` → `0x2002`) |
| 3 | SDK version negotiation via `GetIFVersion` XML command |
| 4 | Device assigns a GUID for the session |
| 5 | SDK commands sent as XML wrapped in binary frames |

**Packet format:** `[2B length LE][2B command LE][payload]`

**SDK command frame:** `[2B length][2B cmd=0x2003][4B total XML length][4B XML offset][XML fragment]`

Large XML payloads are automatically fragmented (max 8000 bytes per packet).

## Available Effects

| Effect | Constant | Description |
|--------|----------|-------------|
| 0 | `EffectNone` | Instant display |
| 1 | `EffectLeftMove` | Slide left |
| 2 | `EffectRightMove` | Slide right |
| 3 | `EffectUpMove` | Slide up |
| 4 | `EffectDownMove` | Slide down |
| 17 | `EffectFadeIn` | Fade in |
| 21 | `EffectLeftScroll` | Continuous left scroll |
| 25 | `EffectRandom` | Random effect |
| 26 | `EffectLeftScrollLoop` | Looping left scroll |
| ... | ... | 30 effects total |

## Device Options

```go
device := huidu.NewDevice("192.168.6.1", 10001,
	huidu.WithTimeout(10 * time.Second),
	huidu.WithHeartbeatInterval(30 * time.Second),
	huidu.WithAutoReconnect(true),
	huidu.WithLogger(log.Default()),
	huidu.WithProgressCallback(func(p huidu.UploadProgress) {
		fmt.Printf("%.1f%%\n", p.Percent)
	}),
)
```

## License

MIT
