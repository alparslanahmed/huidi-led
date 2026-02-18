// Package huidu provides a comprehensive Go library for communicating with
// Huidu full-color LED controller cards via the Huidu SDK 2.0 binary TCP protocol.
//
// # Overview
//
// This library implements the Huidu SDK 2.0 binary TCP protocol in Go.
// It allows you to connect to LED display controller cards over TCP/IP to send
// text, image, video, and clock programs, manage device settings, and transfer files.
//
// # Protocol Architecture
//
// Huidu SDK 2.0 uses a custom binary protocol over TCP:
//
//   - Each TCP packet uses [2B length LE][2B command LE][payload] framing
//   - SDK commands (0x2003) are XML-based with a 12-byte header
//   - Large XML data is fragmented (max 8000 bytes per fragment)
//   - A 3-phase handshake is performed on connection
//
// # Connection Flow
//
//  1. TCP connection is established (default port: 10001)
//  2. Transport version negotiation (0x2001 -> 0x2002)
//  3. SDK version negotiation (GetIFVersion XML command)
//  4. Device assigns a GUID used in all subsequent commands
//
// # Quick Start
//
//	device := huidu.NewDevice("192.168.6.1", 10001)
//	err := device.Connect()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer device.Close()
//
//	info, err := device.GetDeviceInfo()
//	screen := huidu.NewScreen()
//	prog := screen.AddProgram("test")
//	area := prog.AddArea(0, 0, 64, 32)
//	area.AddText("Hello!", huidu.TextConfig{Color: "#ff0000", FontSize: 14})
//	err = device.SendScreen(screen)
//
// # Supported Features
//
//   - Device info queries (CPU, model, screen size, firmware version)
//   - Text, image, video, and clock programs with 30 transition effects
//   - Brightness management (manual, scheduled, sensor-based)
//   - Screen on/off and scheduled switch control
//   - Ethernet and WiFi network configuration
//   - Time synchronization
//   - File upload (image, video, font, firmware) with resume support
//   - File listing and deletion
//   - Heartbeat-based connection keep-alive
//
// # Thread Safety
//
// The Device struct is thread-safe. All public methods are protected by mutex.
// A single Device instance can be safely used from multiple goroutines.
package huidu
