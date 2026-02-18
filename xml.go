package huidu

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// ─── XML Oluşturma ──────────────────────────────────────────────────────────────
//
// Bu dosya, Huidu SDK protokolünde kullanılan XML yapılarının oluşturulması
// ve ayrıştırılması için fonksiyonları içerir.
//
// SDK XML Formatı:
//
// İstek:
//   <sdk guid="GUID-DEĞER">
//     <in method="MetotAdı">
//       ... alt elemanlar ...
//     </in>
//   </sdk>
//
// Yanıt:
//   <sdk guid="GUID-DEĞER">
//     <out method="MetotAdı" result="kSuccess">
//       ... sonuç verileri ...
//     </out>
//   </sdk>

// buildSdkXML, verilen GUID, metot adı ve iç XML içeriğinden
// SDK istek XML'i oluşturur.
//
// Oluşturulan XML formatı:
//
//	<?xml version="1.0" encoding="utf-8"?>
//	<sdk guid="..."><in method="...">innerContent</in></sdk>
//
// innerXML parametresi boş bırakılabilir (basit komutlar için).
func buildSdkXML(guid string, method SdkMethod, innerXML string) string {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	buf.WriteString("\r\n")
	buf.WriteString(fmt.Sprintf(`<sdk guid="%s">`, xmlEscape(guid)))
	buf.WriteString("\r\n  ")
	buf.WriteString(fmt.Sprintf(`<in method="%s">`, string(method)))
	if innerXML != "" {
		buf.WriteString("\r\n    ")
		buf.WriteString(innerXML)
		buf.WriteString("\r\n  ")
	}
	buf.WriteString(`</in>`)
	buf.WriteString("\r\n")
	buf.WriteString(`</sdk>`)
	return buf.String()
}

// buildVersionXML, SDK versiyon anlaşma XML'i oluşturur.
// Bu, binary handshake'ten sonra gönderilen ilk SDK komutudur.
//
// Format:
//
//	<sdk guid="##GUID"><in method="GetIFVersion">
//	  <version value="1000000"/>
//	</in></sdk>
func buildVersionXML() string {
	versionHex := fmt.Sprintf("%x", sdkVersion)
	inner := fmt.Sprintf(`<version value="%s"/>`, versionHex)
	return buildSdkXML("##GUID", MethodGetIFVersion, inner)
}

// xmlElement, basit bir XML elemanı oluşturmak için yardımcı fonksiyondur.
// Attribute'ları key=value çiftleri olarak alır.
//
// Örnek:
//
//	xmlElement("server", "host", "192.168.1.250", "port", "10001")
//	// Sonuç: <server host="192.168.1.250" port="10001"/>
func xmlElement(tag string, attrs ...string) string {
	var buf bytes.Buffer
	buf.WriteString("<")
	buf.WriteString(tag)
	for i := 0; i+1 < len(attrs); i += 2 {
		buf.WriteString(fmt.Sprintf(` %s="%s"`, attrs[i], xmlEscape(attrs[i+1])))
	}
	buf.WriteString("/>")
	return buf.String()
}

// xmlElementWithContent, içeriği olan bir XML elemanı oluşturur.
//
// Örnek:
//
//	xmlElementWithContent("string", "Hello World")
//	// Sonuç: <string>Hello World</string>
func xmlElementWithContent(tag, content string, attrs ...string) string {
	var buf bytes.Buffer
	buf.WriteString("<")
	buf.WriteString(tag)
	for i := 0; i+1 < len(attrs); i += 2 {
		buf.WriteString(fmt.Sprintf(` %s="%s"`, attrs[i], xmlEscape(attrs[i+1])))
	}
	buf.WriteString(">")
	buf.WriteString(xmlEscape(content))
	buf.WriteString("</")
	buf.WriteString(tag)
	buf.WriteString(">")
	return buf.String()
}

// xmlElementWithChildren, alt elementleri olan bir XML elemanı oluşturur.
//
// Örnek:
//
//	xmlElementWithChildren("eth", []string{"valid", "true"},
//	  xmlElement("enable", "value", "true"),
//	  xmlElement("dhcp", "auto", "true"),
//	)
func xmlElementWithChildren(tag string, attrs []string, children ...string) string {
	var buf bytes.Buffer
	buf.WriteString("<")
	buf.WriteString(tag)
	for i := 0; i+1 < len(attrs); i += 2 {
		buf.WriteString(fmt.Sprintf(` %s="%s"`, attrs[i], xmlEscape(attrs[i+1])))
	}
	if len(children) == 0 {
		buf.WriteString("/>")
		return buf.String()
	}
	buf.WriteString(">")
	for _, child := range children {
		buf.WriteString(child)
	}
	buf.WriteString("</")
	buf.WriteString(tag)
	buf.WriteString(">")
	return buf.String()
}

// ─── XML Ayrıştırma ─────────────────────────────────────────────────────────────

// SdkResponse, cihazdan gelen SDK XML yanıtını temsil eder.
type SdkResponse struct {
	// GUID, bu oturumun benzersiz kimliğidir.
	GUID string

	// Method, çağrılan SDK metot adıdır (ör: "GetDeviceInfo").
	Method string

	// Result, işlem sonucudur (ör: "kSuccess", "kParseXmlFailed").
	Result string

	// InnerXML, <out> elementinin ham iç XML içeriğidir.
	// Ayrıntılı veri ayrıştırma için kullanılır.
	InnerXML string

	// RawXML, yanıtın tamamının ham XML metnidir (debug için).
	RawXML string
}

// IsSuccess, yanıtın başarılı olup olmadığını kontrol eder.
func (r *SdkResponse) IsSuccess() bool {
	return r.Result == "kSuccess"
}

// parseSdkResponse, ham XML yanıtını SdkResponse yapısına ayrıştırır.
//
// Beklenen format:
//
//	<sdk guid="..."><out method="..." result="...">...</out></sdk>
//
// Eğer XML bu formata uymuyorsa hata döner.
func parseSdkResponse(rawXML string) (*SdkResponse, error) {
	resp := &SdkResponse{
		RawXML: rawXML,
	}

	decoder := xml.NewDecoder(strings.NewReader(rawXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "sdk":
				for _, attr := range se.Attr {
					if attr.Name.Local == "guid" {
						resp.GUID = attr.Value
					}
				}
			case "out":
				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "method":
						resp.Method = attr.Value
					case "result":
						resp.Result = attr.Value
					}
				}
				// İç XML'i oku
				var innerBuf bytes.Buffer
				depth := 1
				for depth > 0 {
					innerTok, err := decoder.Token()
					if err != nil {
						break
					}
					switch t := innerTok.(type) {
					case xml.StartElement:
						depth++
						innerBuf.WriteString("<")
						innerBuf.WriteString(t.Name.Local)
						for _, a := range t.Attr {
							innerBuf.WriteString(fmt.Sprintf(` %s="%s"`, a.Name.Local, xmlEscape(a.Value)))
						}
						innerBuf.WriteString(">")
					case xml.EndElement:
						depth--
						if depth > 0 {
							innerBuf.WriteString("</")
							innerBuf.WriteString(t.Name.Local)
							innerBuf.WriteString(">")
						}
					case xml.CharData:
						innerBuf.WriteString(string(t))
					}
				}
				resp.InnerXML = strings.TrimSpace(innerBuf.String())
			}
		}
	}

	if resp.GUID == "" && resp.Method == "" {
		return nil, fmt.Errorf("geçersiz SDK XML yanıtı: GUID veya method bulunamadı")
	}

	return resp, nil
}

// extractGUID, SDK XML verisinden GUID değerini çıkarır.
// Hızlı GUID çıkarma için tam XML ayrıştırma yerine string parsing kullanır.
func extractGUID(xmlData string) string {
	idx := strings.Index(xmlData, `guid="`)
	if idx < 0 {
		return ""
	}
	start := idx + 6
	end := strings.Index(xmlData[start:], `"`)
	if end < 0 {
		return ""
	}
	return xmlData[start : start+end]
}

// replaceGUID, XML verisindeki GUID değerini yenisiyle değiştirir.
// SDK handshake sonrasında "##GUID" placeholder'ını gerçek GUID ile değiştirmek
// için kullanılır.
func replaceGUID(xmlData, oldGUID, newGUID string) string {
	return strings.Replace(xmlData, `guid="`+oldGUID+`"`, `guid="`+newGUID+`"`, 1)
}

// ─── Yanıt Veri Ayrıştırıcılar ─────────────────────────────────────────────────

// parseDeviceInfoXML, GetDeviceInfo yanıtının iç XML'inden DeviceInfo çıkarır.
//
// Beklenen format:
//
//	<device cpu="..." model="..." id="..." name="...">
//	<version fpga="..." app="..." kernel="...">
//	<screen width="..." height="..." rotation="...">
func parseDeviceInfoXML(innerXML string) (*DeviceInfo, error) {
	info := &DeviceInfo{}
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "device":
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "cpu":
					info.CPU = a.Value
				case "model":
					info.Model = a.Value
				case "id":
					info.DeviceID = a.Value
				case "name":
					info.DeviceName = a.Value
				}
			}
		case "version":
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "fpga":
					info.FPGAVersion = a.Value
				case "app":
					info.AppVersion = a.Value
				case "kernel":
					info.KernelVersion = a.Value
				}
			}
		case "screen":
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "width":
					fmt.Sscanf(a.Value, "%d", &info.ScreenWidth)
				case "height":
					fmt.Sscanf(a.Value, "%d", &info.ScreenHeight)
				case "rotation":
					fmt.Sscanf(a.Value, "%d", &info.ScreenRotation)
				}
			}
		}
	}
	return info, nil
}

// parseEthernetInfoXML, GetEth0Info yanıtının iç XML'inden EthernetInfo çıkarır.
func parseEthernetInfoXML(innerXML string) (*EthernetInfo, error) {
	info := &EthernetInfo{}
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "enable":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.Enabled = strings.ToLower(a.Value) == "true"
				}
			}
		case "dhcp":
			for _, a := range se.Attr {
				if a.Name.Local == "auto" {
					info.AutoDHCP = strings.ToLower(a.Value) == "true"
				}
			}
		case "address":
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "ip":
					info.IP = a.Value
				case "netmask":
					info.Netmask = a.Value
				case "gateway":
					info.Gateway = a.Value
				case "dns":
					info.DNS = a.Value
				}
			}
		}
	}
	return info, nil
}

// parseWifiInfoXML, GetWifiInfo yanıtının iç XML'inden WifiInfo çıkarır.
func parseWifiInfoXML(innerXML string) (*WifiInfo, error) {
	info := &WifiInfo{}
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "wifi":
			for _, a := range se.Attr {
				if a.Name.Local == "valid" {
					info.HasWifi = strings.ToLower(a.Value) == "true"
				}
			}
		case "enable":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.Enabled = strings.ToLower(a.Value) == "true"
				}
			}
		case "mode":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					if a.Value == "ap" {
						info.WorkMode = 0
					} else {
						info.WorkMode = 1
					}
				}
			}
		case "ap":
			// AP alt elementleri ayrıca okunacak
		case "ssid":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.APInfo.SSID = a.Value
				}
			}
		case "passwd":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.APInfo.Password = a.Value
				}
			}
		case "channel":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.APInfo.Channel = a.Value
				}
			}
		case "encryption":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.APInfo.Encryption = a.Value
				}
			}
		}
	}
	return info, nil
}

// parseLuminanceInfoXML, GetLuminancePloy yanıtının iç XML'inden LuminanceInfo çıkarır.
func parseLuminanceInfoXML(innerXML string) (*LuminanceInfo, error) {
	info := &LuminanceInfo{DefaultValue: 100, SensorMin: 1, SensorMax: 100, SensorTime: 10}
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "mode":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					switch a.Value {
					case "default":
						info.Mode = 0
					case "ploys":
						info.Mode = 1
					case "sensor":
						info.Mode = 2
					}
				}
			}
		case "default":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					fmt.Sscanf(a.Value, "%d", &info.DefaultValue)
				}
			}
		case "item":
			item := LuminanceItem{Enabled: true, Start: "00:00:00", Percent: 100}
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "enable":
					item.Enabled = strings.ToLower(a.Value) == "true"
				case "start":
					item.Start = a.Value
				case "percent":
					fmt.Sscanf(a.Value, "%d", &item.Percent)
				}
			}
			info.CustomItems = append(info.CustomItems, item)
		case "sensor":
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "min":
					fmt.Sscanf(a.Value, "%d", &info.SensorMin)
				case "max":
					fmt.Sscanf(a.Value, "%d", &info.SensorMax)
				case "time":
					fmt.Sscanf(a.Value, "%d", &info.SensorTime)
				}
			}
		}
	}
	return info, nil
}

// parseTimeInfoXML, GetTimeInfo yanıtının iç XML'inden TimeInfo çıkarır.
func parseTimeInfoXML(innerXML string) (*TimeInfo, error) {
	info := &TimeInfo{}
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "timezone":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.Timezone = a.Value
				}
			}
		case "summer":
			for _, a := range se.Attr {
				if a.Name.Local == "enable" {
					info.Summer = strings.ToLower(a.Value) == "true"
				}
			}
		case "sync":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.Sync = a.Value
				}
			}
		case "time":
			for _, a := range se.Attr {
				if a.Name.Local == "value" {
					info.Time = a.Value
				}
			}
		}
	}
	return info, nil
}

// parseSwitchTimeInfoXML, GetSwitchTime yanıtının iç XML'inden SwitchTimeInfo çıkarır.
func parseSwitchTimeInfoXML(innerXML string) (*SwitchTimeInfo, error) {
	info := &SwitchTimeInfo{OpenEnabled: true}
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "open":
			for _, a := range se.Attr {
				if a.Name.Local == "enable" {
					info.OpenEnabled = strings.ToLower(a.Value) == "true"
				}
			}
		case "ploy":
			for _, a := range se.Attr {
				if a.Name.Local == "enable" {
					info.PloyEnabled = strings.ToLower(a.Value) == "true"
				}
			}
		case "item":
			item := SwitchTimeItem{Enabled: true}
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "enable":
					item.Enabled = strings.ToLower(a.Value) == "true"
				case "start":
					item.Start = a.Value
				case "end":
					item.End = a.Value
				}
			}
			info.Items = append(info.Items, item)
		}
	}
	return info, nil
}

// parseBootLogoInfoXML, GetBootLogo yanıtının iç XML'inden BootLogoInfo çıkarır.
func parseBootLogoInfoXML(innerXML string) (*BootLogoInfo, error) {
	info := &BootLogoInfo{}
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "logo" {
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "exist":
					info.Exists = strings.ToLower(a.Value) == "true"
				case "name":
					info.Name = a.Value
				case "md5":
					info.MD5 = a.Value
				}
			}
		}
	}
	return info, nil
}

// parseFontInfoXML, GetAllFontInfo yanıtının iç XML'inden FontInfo listesi çıkarır.
func parseFontInfoXML(innerXML string) ([]FontInfo, error) {
	var fonts []FontInfo
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "font" {
			f := FontInfo{}
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "fontName":
					f.FontName = a.Value
				case "fileName":
					f.FileName = a.Value
				case "bold":
					f.Bold = strings.ToLower(a.Value) == "true"
				case "italic":
					f.Italic = strings.ToLower(a.Value) == "true"
				case "underline":
					f.Underline = strings.ToLower(a.Value) == "true"
				}
			}
			fonts = append(fonts, f)
		}
	}
	return fonts, nil
}

// parseServerInfoXML, GetSDKTcpServer yanıtının iç XML'inden ServerInfo çıkarır.
func parseServerInfoXML(innerXML string) (*ServerInfo, error) {
	info := &ServerInfo{Port: DefaultPort}
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "server" {
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "host":
					info.Host = a.Value
				case "port":
					fmt.Sscanf(a.Value, "%d", &info.Port)
				}
			}
		}
	}
	return info, nil
}

// parseFileListXML, GetFiles yanıtının iç XML'inden FileInfo listesi çıkarır.
func parseFileListXML(innerXML string) ([]FileInfo, error) {
	var files []FileInfo
	decoder := xml.NewDecoder(strings.NewReader(innerXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local == "file" {
			f := FileInfo{}
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "name":
					f.Name = a.Value
				case "size":
					fmt.Sscanf(a.Value, "%d", &f.Size)
				case "existSize":
					fmt.Sscanf(a.Value, "%d", &f.ExistSize)
				case "md5":
					f.MD5 = a.Value
				case "type":
					f.Type = a.Value
				}
			}
			files = append(files, f)
		}
	}
	return files, nil
}

// ─── XML Yardımcı Fonksiyonlar ──────────────────────────────────────────────────

// xmlEscape, XML özel karakterlerini güvenli formata dönüştürür.
func xmlEscape(s string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}

// cleanXML, C# SDK'nın ürettiği XML'deki BOM karakterlerini temizler.
// UTF-8 BOM (0xEF, 0xBB, 0xBF) ile başlayan XML'ler Go'nun xml
// ayrıştırıcısında sorun çıkarabilir.
func cleanXML(data []byte) string {
	// UTF-8 BOM kaldır
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}
	return strings.TrimSpace(string(data))
}

// ─── Set Komutları İçin XML Oluşturucular ───────────────────────────────────────

// buildSetEthernetXML, SetEth0Info komutunun XML içeriğini oluşturur.
func buildSetEthernetXML(info *EthernetInfo) string {
	enableElem := xmlElement("enable", "value", boolStr(info.Enabled))
	dhcpElem := xmlElement("dhcp", "auto", boolStr(info.AutoDHCP))
	addrElem := xmlElement("address",
		"ip", info.IP,
		"netmask", info.Netmask,
		"gateway", info.Gateway,
		"dns", info.DNS,
	)
	return xmlElementWithChildren("eth", []string{"valid", "true"}, enableElem, dhcpElem, addrElem)
}

// buildSetLuminanceXML, SetLuminancePloy komutunun XML içeriğini oluşturur.
func buildSetLuminanceXML(info *LuminanceInfo) string {
	modeStr := "default"
	switch info.Mode {
	case 1:
		modeStr = "ploys"
	case 2:
		modeStr = "sensor"
	}

	parts := []string{
		xmlElement("mode", "value", modeStr),
		xmlElement("default", "value", fmt.Sprintf("%d", info.DefaultValue)),
	}

	// Zamanlı parlaklık öğeleri
	var ployChildren []string
	for _, item := range info.CustomItems {
		ployChildren = append(ployChildren,
			xmlElement("item",
				"enable", boolStr(item.Enabled),
				"start", item.Start,
				"percent", fmt.Sprintf("%d", item.Percent),
			))
	}
	parts = append(parts, xmlElementWithChildren("ploy", nil, ployChildren...))

	// Sensör ayarları
	parts = append(parts, xmlElement("sensor",
		"min", fmt.Sprintf("%d", info.SensorMin),
		"max", fmt.Sprintf("%d", info.SensorMax),
		"time", fmt.Sprintf("%d", info.SensorTime),
	))

	return strings.Join(parts, "")
}

// buildSetTimeXML, SetTimeInfo komutunun XML içeriğini oluşturur.
func buildSetTimeXML(info *TimeInfo) string {
	return strings.Join([]string{
		xmlElement("timezone", "value", info.Timezone),
		xmlElement("summer", "enable", boolStr(info.Summer)),
		xmlElement("sync", "value", info.Sync),
		xmlElement("time", "value", info.Time),
	}, "")
}

// buildSetSwitchTimeXML, SetSwitchTime komutunun XML içeriğini oluşturur.
func buildSetSwitchTimeXML(info *SwitchTimeInfo) string {
	openElem := xmlElement("open", "enable", boolStr(info.OpenEnabled))

	var ployChildren []string
	for _, item := range info.Items {
		ployChildren = append(ployChildren,
			xmlElement("item",
				"enable", boolStr(item.Enabled),
				"start", item.Start,
				"end", item.End,
			))
	}
	ployElem := xmlElementWithChildren("ploy",
		[]string{"enable", boolStr(info.PloyEnabled)},
		ployChildren...,
	)

	return openElem + ployElem
}

// buildSetServerXML, SetSDKTcpServer komutunun XML içeriğini oluşturur.
func buildSetServerXML(info *ServerInfo) string {
	return xmlElement("server", "host", info.Host, "port", fmt.Sprintf("%d", info.Port))
}

// buildDeleteFilesXML, DeleteFiles komutunun XML içeriğini oluşturur.
func buildDeleteFilesXML(fileNames []string) string {
	var children []string
	for _, name := range fileNames {
		children = append(children, xmlElement("file", "name", name))
	}
	return xmlElementWithChildren("files", nil, children...)
}

// buildSetBootLogoXML, SetBootLogoName komutunun XML içeriğini oluşturur.
func buildSetBootLogoXML(info *BootLogoInfo) string {
	return xmlElement("logo",
		"exist", boolStr(info.Exists),
		"name", info.Name,
		"md5", info.MD5,
	)
}

// boolStr, Go bool değerini SDK'nın beklediği lowercase string'e dönüştürür.
func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
