# huidi-led

[![Go Reference](https://pkg.go.dev/badge/github.com/alparslanahmed/huidi-led.svg)](https://pkg.go.dev/github.com/alparslanahmed/huidi-led)
[![Go Report Card](https://goreportcard.com/badge/github.com/alparslanahmed/huidi-led)](https://goreportcard.com/report/github.com/alparslanahmed/huidi-led)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A comprehensive Go library for communicating with **Huidu full-color LED controller cards** via the Huidu SDK 2.0 binary TCP protocol.

> Reverse-engineered from the official Huidu C# SDK 2.0.10. Fully tested with Huidu HD-WF2 and compatible controllers.

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
  - [Device Connection](#device-connection)
  - [Device Info](#device-info)
  - [Display Programs](#display-programs)
  - [Text Display](#text-display)
  - [Scrolling Text](#scrolling-text)
  - [Digital Clock](#digital-clock)
  - [Image Display](#image-display)
  - [Video Display](#video-display)
  - [Multi-Program Screen](#multi-program-screen)
  - [Program Management](#program-management)
  - [Brightness Control](#brightness-control)
  - [Screen On/Off](#screen-onoff)
  - [Network Configuration](#network-configuration)
  - [Time Sync](#time-sync)
  - [File Management](#file-management)
  - [Boot Logo](#boot-logo)
  - [TCP Server Settings](#tcp-server-settings)
  - [Raw XML Commands](#raw-xml-commands)
- [Transition Effects](#transition-effects)
- [Color Constants](#color-constants)
- [Configuration Options](#configuration-options)
- [Protocol Details](#protocol-details)
- [Data Types](#data-types)
- [Multi-Panel Setup](#multi-panel-setup)
- [Error Handling](#error-handling)
- [Thread Safety](#thread-safety)
- [Changelog](#changelog)
- [License](#license)

---

## Features

| Category | Capabilities |
|----------|-------------|
| **Device Management** | Query device info (CPU, model, screen size, firmware, FPGA, kernel version) |
| **Display Programs** | Send text, image, video, and clock programs with 30+ transition effects |
| **Screen Builder** | Hierarchical Screen -> Program -> Area -> Item builder API |
| **Brightness Control** | Manual percentage, scheduled time-based, and sensor-based brightness |
| **Screen Control** | Immediate open/close screen, scheduled on/off timers |
| **Network Config** | Ethernet (IP/DHCP/DNS) and WiFi (AP & Station mode) configuration |
| **Time Sync** | Synchronize device clock, timezone and DST support |
| **File Transfer** | Upload images, videos, fonts, firmware with MD5 verification, chunked transfer, resume support, and progress callbacks |
| **File Management** | List files on device, delete single or multiple files |
| **Boot Logo** | Get/set/clear boot logo image |
| **TCP Server** | Configure remote TCP server settings |
| **Heartbeat** | Automatic connection keep-alive with configurable interval |
| **Thread-Safe** | All public methods are mutex-protected; safe for concurrent goroutines |
| **Auto Screen Size** | Queries device for actual screen dimensions on connect |

---

## Installation

```bash
go get github.com/alparslanahmed/huidi-led
```

Requires Go 1.21 or later.

---

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "time"

    huidu "github.com/alparslanahmed/huidi-led"
)

func main() {
    // Create device with options
    device := huidu.NewDevice("192.168.6.1", 10001,
        huidu.WithTimeout(10*time.Second),
        huidu.WithLogger(log.Default()),
    )

    // Connect (performs 3-phase handshake automatically)
    if err := device.Connect(); err != nil {
        log.Fatal(err)
    }
    defer device.Close()

    // Screen size is auto-detected from device
    info := device.CachedDeviceInfo()
    fmt.Printf("Connected: %s, Screen: %dx%d\n",
        info.DeviceID, info.ScreenWidth, info.ScreenHeight)

    // Send a red text message
    screen := huidu.NewScreen()
    prog := screen.AddProgram("hello")
    area := prog.AddFullScreenArea(info.ScreenWidth, info.ScreenHeight)
    area.AddText("Hello World!", huidu.TextConfig{
        Color:    huidu.ColorRed,
        FontSize: 14,
        Effect:   huidu.EffectLeftScrollLoop,
        Speed:    4,
    })

    if err := device.SendScreen(screen); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Text sent successfully!")
}
```

---

## API Reference

### Device Connection

```go
// Create a new device
device := huidu.NewDevice(host, port, options...)

// Connect to device (3-phase handshake: version -> SDK version -> device info)
err := device.Connect()

// Disconnect
err := device.Close()

// Check connection status
connected := device.IsConnected()

// Get connection details
host := device.Host()
port := device.Port()
guid := device.GUID()             // Session GUID assigned by device
info := device.CachedDeviceInfo()  // Device info from last handshake
```

### Device Info

```go
// Query full device information
info, err := device.GetDeviceInfo()

// DeviceInfo fields:
//   DeviceID       string  - Unique device identifier
//   DeviceName     string  - User-assigned device name
//   Model          string  - Hardware model (e.g. "HD-WF2")
//   CPU            string  - CPU model
//   AppVersion     string  - Firmware version
//   FPGAVersion    string  - FPGA firmware version
//   KernelVersion  string  - Linux kernel version
//   ScreenWidth    int     - Total screen width in pixels
//   ScreenHeight   int     - Total screen height in pixels
//   ScreenRotation int     - Rotation angle (0, 90, 180, 270)

fmt.Printf("Device: %s (%s)\n", info.DeviceID, info.Model)
fmt.Printf("Screen: %dx%d, Firmware: %s\n",
    info.ScreenWidth, info.ScreenHeight, info.AppVersion)
```

### Display Programs

The display hierarchy is: **Screen -> Program(s) -> Area(s) -> Item(s)**

```
Screen (root container)
+-- Program 1 (plays in sequence)
|   +-- Area A (region on screen)
|   |   +-- Text item
|   |   +-- Image item
|   +-- Area B
|       +-- Clock item
+-- Program 2
    +-- Area C
        +-- Video item
```

```go
// Create a screen (replaces ALL content on device)
screen := huidu.NewScreen()

// Add a program to the screen
prog := screen.AddProgram("Program Name")

// Or with advanced config
prog := screen.AddProgramWithConfig(huidu.ProgramConfig{
    Name:      "Emergency",
    Type:      huidu.ProgramNormal,
    PlayCount: 3,           // Play 3 times then move to next program
    Duration:  "00:05:00",  // Or play for 5 minutes
    Realtime:  true,        // Realtime flag
    Disabled:  false,       // Set true to skip this program
})

// Add an area (display region) to the program
area := prog.AddArea(x, y, width, height)

// Or a full-screen area
area := prog.AddFullScreenArea(screenWidth, screenHeight)

// Send screen to device (replaces all existing programs)
err := device.SendScreen(screen)
```

### Text Display

```go
screen := huidu.NewScreen()
prog := screen.AddProgram("text_demo")
area := prog.AddFullScreenArea(128, 64)

area.AddText("Hello World!", huidu.TextConfig{
    // Font settings
    FontName: "Arial",       // Font name (must exist on device)
    FontSize: 14,            // Font size in pixels
    Bold:     true,          // Bold text
    Italic:   false,         // Italic text
    Color:    "#ff0000",     // Text color (hex RGB)

    // Alignment
    HAlign: huidu.HAlignCenter,  // "left", "center", "right"
    VAlign: huidu.VAlignMiddle,  // "top", "middle", "bottom"

    // Transition effect
    Effect:   huidu.EffectLeftScrollLoop,  // See effects table below
    Speed:    4,                            // 1-10 (higher = faster)
    Duration: 30,                           // Display duration in seconds

    // Background (optional)
    BackgroundColor: "#000000",  // Background color
    BackgroundImage: "",         // Background image filename
})

device.SendScreen(screen)
```

**Shortcut method** (auto-detects screen size from device):

```go
err := device.SendText("Quick message!", huidu.TextConfig{
    Color:    huidu.ColorGreen,
    FontSize: 14,
    Effect:   huidu.EffectFadeIn,
    Speed:    4,
})
```

### Scrolling Text

```go
screen := huidu.NewScreen()
prog := screen.AddProgram("ticker")
area := prog.AddFullScreenArea(128, 64)

area.AddText("Breaking News: Go library released!", huidu.TextConfig{
    Color:    huidu.ColorYellow,
    FontSize: 14,
    Bold:     true,
    Effect:   huidu.EffectLeftScrollLoop,  // Continuous scrolling
    Speed:    3,
    HAlign:   huidu.HAlignLeft,
    VAlign:   huidu.VAlignMiddle,
})

device.SendScreen(screen)
```

### Digital Clock

```go
screen := huidu.NewScreen()
prog := screen.AddProgram("clock")
area := prog.AddFullScreenArea(128, 64)

area.AddClock(huidu.ClockConfig{
    Type: huidu.ClockDigital,  // "digital" or "analog"

    // Date display
    ShowDate:   true,
    DateColor:  "#ffff00",
    DateFormat: 1,           // 0=YYYY-MM-DD, 1=DD/MM/YYYY, etc.

    // Time display
    ShowTime:   true,
    TimeColor:  "#00ffff",
    TimeFormat: 1,           // 0=24h, 1=12h

    // Week display
    ShowWeek:  true,
    WeekColor: "#00ff00",

    // Title
    ShowTitle:  false,
    TitleColor: "#ff0000",
    TitleText:  "Office Clock",

    // Lunar calendar (Chinese)
    ShowLunarCalendar:  false,
    LunarCalendarColor: "#ff0000",
})

device.SendScreen(screen)
```

### Image Display

```go
// First upload the image to the device
device.UploadFile("/path/to/logo.jpg")

// Then display it
screen := huidu.NewScreen()
prog := screen.AddProgram("image_demo")
area := prog.AddFullScreenArea(128, 64)

area.AddImage("logo.jpg", huidu.ImageConfig{
    Fit:      huidu.ImageFitStretch,  // "stretch", "center", "fill"
    Duration: 10,                      // Display for 10 seconds
    Effect:   huidu.EffectFadeIn,
    Speed:    3,
})

device.SendScreen(screen)
```

### Video Display

```go
// Upload video first
device.UploadFile("/path/to/clip.mp4")

// Display it
screen := huidu.NewScreen()
prog := screen.AddProgram("video_demo")
area := prog.AddFullScreenArea(128, 64)

area.AddVideo("clip.mp4", huidu.VideoConfig{
    Volume: 50,  // 0-100
})

device.SendScreen(screen)
```

### Multi-Program Screen

Programs are played in sequence. You can combine different content types:

```go
screen := huidu.NewScreen()

// Program 1: Welcome text for 10 seconds
prog1 := screen.AddProgramWithConfig(huidu.ProgramConfig{
    Name:     "Welcome",
    Duration: "00:00:10",
})
area1 := prog1.AddFullScreenArea(128, 64)
area1.AddText("Welcome!", huidu.TextConfig{
    Color: huidu.ColorWhite, FontSize: 16, Effect: huidu.EffectFadeIn,
})

// Program 2: Scrolling news
prog2 := screen.AddProgramWithConfig(huidu.ProgramConfig{
    Name:      "News",
    PlayCount: 2, // Play twice then move to next
})
area2 := prog2.AddFullScreenArea(128, 64)
area2.AddText("Today's headlines...", huidu.TextConfig{
    Color: huidu.ColorYellow, FontSize: 14,
    Effect: huidu.EffectLeftScrollLoop, Speed: 3,
})

// Program 3: Clock (continuous)
prog3 := screen.AddProgram("Clock")
area3 := prog3.AddFullScreenArea(128, 64)
area3.AddClock(huidu.ClockConfig{
    Type: huidu.ClockDigital, ShowTime: true, TimeColor: "#00ffff",
})

// Send all programs - they rotate automatically
device.SendScreen(screen)
```

### Program Management

```go
// Delete ALL programs (clears screen)
err := device.DeleteAllPrograms()

// Update a specific program (must match existing GUID)
err := device.UpdateProgram(program)

// Delete a specific program
err := device.DeleteProgram(program)
```

### Brightness Control

```go
// Quick brightness set (1-100%)
err := device.SetBrightness(80)

// Get current brightness info
info, err := device.GetLuminanceInfo()
fmt.Printf("Mode: %d, Default: %d%%\n", info.Mode, info.DefaultValue)
fmt.Printf("Sensor range: %d%% - %d%%\n", info.SensorMin, info.SensorMax)

// Set full luminance configuration
err := device.SetLuminanceInfo(&huidu.LuminanceInfo{
    Mode:         1,  // Scheduled mode
    DefaultValue: 60,
    CustomItems: []huidu.LuminanceItem{
        {Start: "06:00", Percent: 80, Enabled: true},
        {Start: "18:00", Percent: 40, Enabled: true},
        {Start: "22:00", Percent: 20, Enabled: true},
    },
})
```

### Screen On/Off

```go
// Immediate control
err := device.OpenScreen()
err := device.CloseScreen()

// Scheduled on/off
info, err := device.GetSwitchTimeInfo()

err := device.SetSwitchTimeInfo(&huidu.SwitchTimeInfo{
    OpenEnabled: true,
    PloyEnabled: true,
    Items: []huidu.SwitchTimeItem{
        {Start: "08:00", End: "23:00", Enabled: true},
    },
})
```

### Network Configuration

```go
// --- Ethernet ---
eth, err := device.GetEthernetInfo()
fmt.Printf("IP: %s, DHCP: %v, Gateway: %s\n", eth.IP, eth.AutoDHCP, eth.Gateway)

err := device.SetEthernetInfo(&huidu.EthernetInfo{
    Enabled:  true,
    AutoDHCP: false,
    IP:       "192.168.1.100",
    Netmask:  "255.255.255.0",
    Gateway:  "192.168.1.1",
    DNS:      "8.8.8.8",
})

// --- WiFi ---
wifi, err := device.GetWifiInfo()
fmt.Printf("WiFi: %v, Mode: %d, AP SSID: %s\n",
    wifi.Enabled, wifi.WorkMode, wifi.APInfo.SSID)

err := device.SetWifiInfo(&huidu.WifiInfo{
    Enabled:  true,
    WorkMode: 0, // 0=AP, 1=Station
    APInfo: huidu.WifiAPInfo{
        SSID:     "MyLED",
        Password: "12345678",
        IP:       "192.168.6.1",
        Channel:  6,
    },
})
```

### Time Sync

```go
// Get current device time
info, err := device.GetTimeInfo()
fmt.Printf("Time: %s, Zone: %s\n", info.Time, info.Timezone)

// Set device time
err := device.SetTimeInfo(&huidu.TimeInfo{
    Time:     "2026-02-18 15:30:00",
    Timezone: "(UTC+03:00)Istanbul",
    Sync:     "none",    // "none", "ntp", "gps"
    Summer:   false,     // Daylight saving time
})
```

### File Management

```go
// Upload a file with progress tracking
device := huidu.NewDevice("192.168.6.1", 10001,
    huidu.WithProgressCallback(func(p huidu.UploadProgress) {
        fmt.Printf("\rUploading: %s [%.1f%%] %d/%d bytes",
            p.FileName, p.Percent, p.SentBytes, p.TotalBytes)
    }),
)

err := device.UploadFile("/path/to/image.jpg")

// List files on device
files, err := device.GetFileList()
for _, f := range files {
    fmt.Printf("%s (size: %d, type: %d)\n", f.Name, f.Size, f.Type)
}

// Delete specific files
err := device.DeleteFiles("image1.jpg", "video.mp4")

// Get font list
fonts, err := device.GetFontInfo()
for _, f := range fonts {
    fmt.Printf("%s (%s)\n", f.FontName, f.FileName)
}
```

### Boot Logo

```go
// Get current boot logo info
logo, err := device.GetBootLogoInfo()

// Set boot logo (file must be uploaded first)
err := device.SetBootLogo(&huidu.BootLogoInfo{FileName: "logo.jpg"})

// Clear boot logo
err := device.ClearBootLogo()
```

### TCP Server Settings

```go
// Get remote server config
server, err := device.GetServerInfo()
fmt.Printf("Server: %s:%d, Enabled: %v\n", server.IP, server.Port, server.Enabled)

// Set remote server
err := device.SetServerInfo(&huidu.ServerInfo{
    Enabled: true,
    IP:      "cloud.example.com",
    Port:    10001,
})
```

### Raw XML Commands

For advanced use cases or unsupported commands:

```go
resp, err := device.SendRawXML("<sdk><in method=\"GetDeviceInfo\"></in></sdk>")
fmt.Printf("Success: %v, Response: %s\n", resp.IsSuccess(), resp.InnerXML)
```

---

## Transition Effects

All 30+ supported transition effects:

| Constant | Value | Description |
|----------|-------|-------------|
| EffectNone | 0 | Instant display (no animation) |
| EffectLeftMove | 1 | Slide in from right to left |
| EffectRightMove | 2 | Slide in from left to right |
| EffectUpMove | 3 | Slide in from bottom to top |
| EffectDownMove | 4 | Slide in from top to bottom |
| EffectLeftExpand | 5 | Expand from left edge |
| EffectRightExpand | 6 | Expand from right edge |
| EffectUpExpand | 7 | Expand from top edge |
| EffectDownExpand | 8 | Expand from bottom edge |
| EffectHCenterExpand | 9 | Expand from horizontal center |
| EffectVCenterExpand | 10 | Expand from vertical center |
| EffectCenterExpand | 11 | Expand from center point |
| EffectLeftDiamond | 12 | Diamond wipe from left |
| EffectRightDiamond | 13 | Diamond wipe from right |
| EffectUpDiamond | 14 | Diamond wipe from top |
| EffectDownDiamond | 15 | Diamond wipe from bottom |
| EffectHShutter | 16 | Horizontal shutter/blinds effect |
| EffectFadeIn | 17 | Gradual fade in |
| EffectHSnow | 18 | Horizontal snow/dissolve |
| EffectVSnow | 19 | Vertical snow/dissolve |
| EffectVShutter | 20 | Vertical shutter/blinds effect |
| EffectLeftScroll | 21 | Continuous scroll left (one pass) |
| EffectRightScroll | 22 | Continuous scroll right (one pass) |
| EffectUpScroll | 23 | Continuous scroll up (one pass) |
| EffectDownScroll | 24 | Continuous scroll down (one pass) |
| EffectRandom | 25 | Random effect (device picks) |
| **EffectLeftScrollLoop** | **26** | **Continuous left scroll (loops)** |
| EffectUpScrollLoop | 27 | Continuous up scroll (loops) |

### Effect Helpers

```go
// Check if an effect is a continuous scrolling type
huidu.EffectLeftScrollLoop.IsContinuousScroll() // true
huidu.EffectFadeIn.IsContinuousScroll()         // false

// Check if vertical scroll
huidu.EffectUpScrollLoop.IsVerticalScroll() // true
```

---

## Color Constants

Pre-defined color constants for convenience:

| Constant | Value | Color |
|----------|-------|-------|
| ColorRed | #ff0000 | Red |
| ColorGreen | #00ff00 | Green |
| ColorBlue | #0000ff | Blue |
| ColorWhite | #ffffff | White |
| ColorYellow | #ffff00 | Yellow |
| ColorCyan | #00ffff | Cyan |
| ColorMagenta | #ff00ff | Magenta |
| ColorOrange | #ff8000 | Orange |

Custom colors:

```go
// Using hex string
config := huidu.TextConfig{Color: "#ff8800"}

// Using RGB helper
config := huidu.TextConfig{Color: huidu.RGB(255, 136, 0)}
```

---

## Configuration Options

Options are passed to NewDevice():

```go
device := huidu.NewDevice("192.168.6.1", 10001,
    // TCP connection timeout (default: 5s)
    huidu.WithTimeout(10 * time.Second),

    // Heartbeat interval to keep connection alive (default: 30s)
    huidu.WithHeartbeatInterval(30 * time.Second),

    // Auto-reconnect on connection loss
    huidu.WithAutoReconnect(true),

    // Logger for debug output
    huidu.WithLogger(log.Default()),

    // File upload progress callback
    huidu.WithProgressCallback(func(p huidu.UploadProgress) {
        fmt.Printf("%.1f%%\n", p.Percent)
    }),
)
```

| Option | Default | Description |
|--------|---------|-------------|
| WithTimeout | 5s | TCP connection and read/write timeout |
| WithHeartbeatInterval | 30s | Keep-alive ping interval |
| WithAutoReconnect | false | Automatically reconnect on disconnect |
| WithLogger | nil | Logger for debug messages |
| WithProgressCallback | nil | Callback for file upload progress |

---

## Protocol Details

This library implements the **Huidu SDK 2.0** binary TCP protocol, reverse-engineered from the official C# SDK.

### Connection Handshake (3 phases)

```
Client                          Device (port 10001)
  |                                |
  |---- TCP Connect -------------->|
  |                                |
  |---- Version Request (0x2001) ->|  Phase 1: Transport version
  |<--- Version Reply (0x2002) ----|
  |                                |
  |---- GetIFVersion XML (0x2003) >|  Phase 2: SDK version
  |<--- GetIFVersion Response -----|  (assigns session GUID)
  |                                |
  |---- GetDeviceInfo XML (0x2003)>|  Phase 3: Device info
  |<--- DeviceInfo Response -------|  (screen size, model, etc.)
  |                                |
  |---- Heartbeat loop starts -----|
```

### Packet Format

Every TCP packet uses little-endian framing:

```
+----------+----------+----------------+
| 2 bytes  | 2 bytes  | N bytes        |
| Length LE | Cmd LE   | Payload        |
+----------+----------+----------------+
```

### SDK Command (0x2003) Frame

```
+----------+----------+--------------+--------------+------------+
| 2B       | 2B       | 4B           | 4B           | Variable   |
| Pkt Len  | 0x2003   | Total XML    | XML Offset   | XML Data   |
|          |          | Length LE     | LE           | Fragment   |
+----------+----------+--------------+--------------+------------+
```

- XML payloads larger than 8000 bytes are automatically fragmented
- Device reassembles fragments using offset and total length fields

### File Upload Protocol

```
Client                               Device
  |                                     |
  |-- FileStartAsk (0x8001) ---------->|  MD5 + file size + type
  |<-- FileStartAsk Response ----------|  Resume offset (if partial)
  |                                     |
  |-- FileContentAsk (0x8003) -------->|  Chunk data (max 8000B)
  |<-- FileContentAsk Response --------|  ACK
  |-- FileContentAsk ... ------------>|  (repeat)
  |                                     |
  |-- FileEndAsk (0x8005) ------------>|  Transfer complete
  |<-- FileEndAsk Response ------------|  Final ACK
```

### Supported Command Types

| Code | Name | Direction |
|------|------|-----------|
| 0x2001 | VersionAsk | Client -> Device |
| 0x2002 | VersionReply | Device -> Client |
| 0x2003 | SdkCmd | Client -> Device |
| 0x2004 | SdkReply | Device -> Client |
| 0x2005 | Heartbeat | Client -> Device |
| 0x2006 | HeartbeatReply | Device -> Client |
| 0x8001 | FileStartAsk | Client -> Device |
| 0x8002 | FileStartReply | Device -> Client |
| 0x8003 | FileContentAsk | Client -> Device |
| 0x8004 | FileContentReply | Device -> Client |
| 0x8005 | FileEndAsk | Client -> Device |
| 0x8006 | FileEndReply | Device -> Client |

### SDK XML Methods

| Method | Description |
|--------|-------------|
| GetIFVersion | SDK protocol version negotiation |
| AddProgram | Send/replace all screen programs |
| UpdateProgram | Update a specific program |
| DeleteProgram | Delete a specific program |
| GetDeviceInfo | Query device information |
| GetEth0Info / SetEth0Info | Ethernet configuration |
| GetWifiInfo / SetWifiInfo | WiFi configuration |
| GetLuminancePloy / SetLuminancePloy | Brightness settings |
| GetSwitchTime / SetSwitchTime | Scheduled on/off |
| OpenScreen / CloseScreen | Immediate screen control |
| GetTimeInfo / SetTimeInfo | Time sync |
| GetAllFontInfo | Query available fonts |
| GetFiles / DeleteFiles | File management |
| GetBootLogo / SetBootLogoName / ClearBootLogo | Boot logo |
| GetSDKTcpServer / SetSDKTcpServer | TCP server config |
| GetProgram | Read back current program |
| GetCurrentPlayProgramGUID | Get currently playing program |

---

## Data Types

### Core Structs

| Type | Description |
|------|-------------|
| Device | Main controller connection; all operations go through this |
| Screen | Root display container; holds programs |
| Program | A playable program; holds areas |
| Area | A rectangular display region; holds items |
| TextConfig | Text item settings (font, color, effect, alignment) |
| ImageConfig | Image item settings (fit mode, effect) |
| VideoConfig | Video item settings (volume) |
| ClockConfig | Clock item settings (format, colors, components) |
| ProgramConfig | Advanced program settings (play count, duration, etc.) |

### Info Structs

| Type | Description |
|------|-------------|
| DeviceInfo | Device metadata (ID, model, CPU, firmware, screen size) |
| EthernetInfo | Ethernet network settings |
| WifiInfo | WiFi settings (AP and Station mode) |
| WifiAPInfo | WiFi Access Point details |
| TimeInfo | Time, timezone, sync settings |
| LuminanceInfo | Brightness configuration |
| LuminanceItem | Scheduled brightness entry |
| SwitchTimeInfo | Screen on/off schedule |
| SwitchTimeItem | Individual schedule entry |
| BootLogoInfo | Boot logo settings |
| ServerInfo | Remote TCP server settings |
| FontInfo | Font name and filename |
| FileInfo | File name, size, and type |
| UploadProgress | File upload progress data |

### Enums

| Type | Values |
|------|--------|
| EffectType | 30 transition effects (see table above) |
| FileType | FileImage, FileVideo, FileFont, FileFirmware, FileOther |
| ProgramType | ProgramNormal, ProgramTemplate, ProgramHTML5, ProgramOffline |
| HAlign | HAlignLeft, HAlignCenter, HAlignRight |
| VAlign | VAlignTop, VAlignMiddle, VAlignBottom |
| ClockType | ClockDigital, ClockAnalog |
| ImageFit | ImageFitStretch, ImageFitCenter, ImageFitFill |

---

## Multi-Panel Setup

If your LED display uses multiple panels (e.g., 4x 64x32 panels arranged in a 2x2 grid = 128x64 total), the library auto-detects the total screen size from the device:

```go
device.Connect()

// After connect, CachedDeviceInfo has the real screen dimensions
info := device.CachedDeviceInfo()
w, h := info.ScreenWidth, info.ScreenHeight  // e.g. 128, 64

// Use the full dimensions for areas
screen := huidu.NewScreen()
prog := screen.AddProgram("fullscreen")
area := prog.AddFullScreenArea(w, h)
area.AddText("Spans all panels!", huidu.TextConfig{
    Color: huidu.ColorWhite, FontSize: 16,
})
device.SendScreen(screen)
```

The SendText() shortcut also auto-detects screen size:

```go
// Automatically uses device-reported screen dimensions
device.SendText("Auto-sized!", huidu.TextConfig{
    Color: "#ff0000", FontSize: 14,
})
```

---

## Error Handling

All methods return Go errors. Device-level errors include descriptive messages:

```go
err := device.Connect()
if err != nil {
    // Connection errors, timeout errors, handshake failures
    log.Fatal(err)
}

err = device.SendScreen(screen)
if err != nil {
    // "SendScreen failed: <device error message>"
    log.Fatal(err)
}
```

The library defines ErrorCode type for protocol-level errors:

```go
const (
    ErrNone         ErrorCode = 0   // Success
    ErrUnknown      ErrorCode = 1   // Unknown error
    ErrTimeout      ErrorCode = 2   // Operation timeout
    ErrInvalidParam ErrorCode = 3   // Invalid parameter
    ErrNoMemory     ErrorCode = 4   // Out of memory
    ErrFileNotFound ErrorCode = 5   // File not found
    // ... and more
)
```

---

## Thread Safety

The Device struct is fully thread-safe. All public methods use internal mutex locking:

```go
device.Connect()

// Safe to call from multiple goroutines
go func() { device.GetDeviceInfo() }()
go func() { device.SetBrightness(80) }()
go func() { device.SendText("Hi", huidu.TextConfig{}) }()
```

---

## Changelog

### v0.1.0 (2026-02-18)

- Initial release
- **Fixed**: DeleteAllPrograms() now correctly clears the screen by sending an empty screen via AddProgram (replaces all content), instead of sending an empty DeleteProgram which had no effect
- Full Huidu SDK 2.0 binary TCP protocol implementation
- 3-phase connection handshake (version, SDK version, device info)
- Screen -> Program -> Area -> Item builder API
- 24 device commands: GetDeviceInfo, brightness, screen on/off, time sync, network config, etc.
- 30 transition effects for text, image, video, and clock items
- File upload with MD5 verification, chunked transfer, resume support, and progress callbacks
- Automatic heartbeat keep-alive
- Thread-safe Device struct
- Auto-detection of screen dimensions from device (multi-panel support)
- Color constants and RGB helper
- Functional options pattern for device configuration

---

## License

MIT License - see [LICENSE](LICENSE) for details.
