package huidu

import (
	"encoding/binary"
)

// ─── Paket Oluşturma ────────────────────────────────────────────────────────────
//
// Bu dosya, Huidu binary TCP protokolü için düşük seviyeli paket oluşturma
// fonksiyonlarını içerir. Tüm paketler little-endian byte sıralaması kullanır.
//
// Paket Genel Formatı:
//   [2 byte] Toplam paket uzunluğu (LE)
//   [2 byte] Komut tipi (LE)
//   [N byte] Veri (komuta göre değişir)

// buildVersionPacket, transport protokol versiyon anlaşma paketi oluşturur.
// Bu, TCP bağlantısı kurulduktan sonra gönderilen ilk pakettir.
//
// Paket Formatı (toplam 8 byte):
//
//	[2B] uzunluk = 0x0008
//	[2B] komut   = 0x2001 (CmdServiceAsk)
//	[4B] versiyon = 0x1000005 (transportVersion)
//
// Cihaz, aynı formatta CmdServiceAnswer (0x2002) ile yanıt verir.
func buildVersionPacket() []byte {
	pkt := make([]byte, 8)
	binary.LittleEndian.PutUint16(pkt[0:2], 8) // length
	binary.LittleEndian.PutUint16(pkt[2:4], uint16(CmdServiceAsk))
	binary.LittleEndian.PutUint32(pkt[4:8], transportVersion)
	return pkt
}

// buildHeartbeat, heartbeat (nabız) paketi oluşturur.
// TCP bağlantısı canlı tutmak için DefaultHeartbeatInterval aralığında gönderilir.
//
// Paket Formatı (toplam 4 byte):
//
//	[2B] uzunluk = 0x0004
//	[2B] komut   = 0x005f (CmdHeartbeatAsk)
//
// Cihaz, CmdHeartbeatAnswer (0x0060) ile yanıt verir.
// Eğer 3 heartbeat aralığı boyunca yanıt gelmezse bağlantı kopmuş kabul edilir.
func buildHeartbeat() []byte {
	pkt := make([]byte, 4)
	binary.LittleEndian.PutUint16(pkt[0:2], 4)
	binary.LittleEndian.PutUint16(pkt[2:4], uint16(CmdHeartbeatAsk))
	return pkt
}

// buildSdkCmdPackets, XML tabanlı SDK komutunu binary paketlere dönüştürür.
// Büyük XML verileri MaxContentLength (8000 byte) limitine göre otomatik
// olarak parçalara bölünür (fragmentation).
//
// Her Parça Formatı (toplam 12 + N byte):
//
//	[2B] uzunluk          = 12 + N (paket toplam boyutu)
//	[2B] komut            = 0x2003 (CmdSdkCmdAsk)
//	[4B] toplam XML boyut = tüm XML'in toplam byte uzunluğu (LE)
//	[4B] XML offset       = bu parçanın XML içindeki başlangıç konumu (LE)
//	[NB] XML verisi       = XML'in ilgili kısmı
//
// Cihaz, tüm parçaları aldıktan sonra XML'i birleştirir ve işler.
// Yanıt aynı formatta CmdSdkCmdAnswer (0x2004) olarak döner.
//
// Örnek: 20000 byte'lık bir XML 3 parçaya bölünür:
//
//	Parça 1: offset=0,     boyut=8000
//	Parça 2: offset=8000,  boyut=8000
//	Parça 3: offset=16000, boyut=4000
func buildSdkCmdPackets(xmlData []byte) [][]byte {
	totalLen := len(xmlData)
	if totalLen == 0 {
		return nil
	}

	var packets [][]byte
	offset := 0

	for offset < totalLen {
		// Bu parçada gönderilecek veri miktarını belirle
		chunkSize := totalLen - offset
		if chunkSize > MaxContentLength {
			chunkSize = MaxContentLength
		}

		// Paket oluştur: 12 byte header + chunk data
		pktLen := sdkCmdHeaderLength + chunkSize
		pkt := make([]byte, pktLen)

		// Header: [2B length][2B cmd][4B totalXmlLen][4B xmlOffset]
		binary.LittleEndian.PutUint16(pkt[0:2], uint16(pktLen))
		binary.LittleEndian.PutUint16(pkt[2:4], uint16(CmdSdkCmdAsk))
		binary.LittleEndian.PutUint32(pkt[4:8], uint32(totalLen))
		binary.LittleEndian.PutUint32(pkt[8:12], uint32(offset))

		// XML verisini kopyala
		copy(pkt[sdkCmdHeaderLength:], xmlData[offset:offset+chunkSize])

		packets = append(packets, pkt)
		offset += chunkSize
	}

	return packets
}

// parsePacketHeader, ham TCP verisinden paket başlığını ayrıştırır.
// En az tcpHeaderLength (4) byte'lık veri gerektirir.
//
// Dönen değerler:
//   - length: Toplam paket uzunluğu (header dahil)
//   - cmdType: Komut tipi
//   - ok: Ayrıştırma başarılı mı
func parsePacketHeader(data []byte) (length uint16, cmdType CmdType, ok bool) {
	if len(data) < tcpHeaderLength {
		return 0, 0, false
	}
	length = binary.LittleEndian.Uint16(data[0:2])
	cmdType = CmdType(binary.LittleEndian.Uint16(data[2:4]))
	return length, cmdType, true
}

// parseSdkCmdHeader, SDK komut paketinin genişletilmiş başlığını ayrıştırır.
// En az sdkCmdHeaderLength (12) byte'lık veri gerektirir.
//
// Dönen değerler:
//   - totalLen: XML'in toplam boyutu (tüm parçalar birleştiğinde)
//   - xmlOffset: Bu parçanın XML içindeki başlangıç konumu
//   - ok: Ayrıştırma başarılı mı
func parseSdkCmdHeader(data []byte) (totalLen uint32, xmlOffset uint32, ok bool) {
	if len(data) < sdkCmdHeaderLength {
		return 0, 0, false
	}
	totalLen = binary.LittleEndian.Uint32(data[4:8])
	xmlOffset = binary.LittleEndian.Uint32(data[8:12])
	return totalLen, xmlOffset, true
}

// parseVersionResponse, CmdServiceAnswer paketinden versiyon numarasını çıkarır.
// Paket en az 8 byte olmalıdır.
func parseVersionResponse(data []byte) (version uint32, ok bool) {
	if len(data) < 8 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(data[4:8]), true
}

// parseErrorCode, CmdErrorAnswer paketinden hata kodunu çıkarır.
// Paket en az 6 byte olmalıdır (4B header + 2B error code).
func parseErrorCode(data []byte) (ErrorCode, bool) {
	if len(data) < 6 {
		return 0, false
	}
	code := binary.LittleEndian.Uint16(data[4:6])
	return ErrorCode(code), true
}

// buildFileStartPacket, dosya transfer başlatma paketi oluşturur.
// C# SDK'daki GetUploadFileStartAsk formatıyla birebir uyumludur.
//
// Paket Formatı (headLen=47):
//
//	[0-1]   length (2B LE)
//	[2-3]   cmd = 0x8001 (CmdFileStartAsk) (2B LE)
//	[4-35]  MD5 hash string (32 bytes UTF-8)
//	[36]    padding
//	[37-40] dosya boyutu (4B LE)
//	[41-44] padding
//	[45-46] dosya tipi (2B LE)
//	[47+]   dosya adı (null-terminated UTF-8 string)
func buildFileStartPacket(fileName string, fileSize int64, fileType FileType, md5Hash string) []byte {
	const headLen = 47
	nameBytes := []byte(fileName)
	pktLen := headLen + len(nameBytes) + 1
	pkt := make([]byte, pktLen)

	// [0-1] Paket uzunluğu
	binary.LittleEndian.PutUint16(pkt[0:2], uint16(pktLen))
	// [2-3] Komut tipi
	binary.LittleEndian.PutUint16(pkt[2:4], uint16(CmdFileStartAsk))
	// [4-35] MD5 hash (32 byte hex string)
	md5Bytes := []byte(md5Hash)
	copy(pkt[4:36], md5Bytes)
	// [37-40] Dosya boyutu
	binary.LittleEndian.PutUint32(pkt[37:41], uint32(fileSize))
	// [45-46] Dosya tipi
	binary.LittleEndian.PutUint16(pkt[45:47], uint16(fileType))
	// [47+] Dosya adı + null terminator
	copy(pkt[headLen:], nameBytes)
	pkt[pktLen-1] = 0

	return pkt
}

// buildFileContentPacket, dosya içeriği paketi oluşturur.
//
// Paket Formatı:
//
//	[2B] length
//	[2B] cmd = 0x8003 (CmdFileContentAsk)
//	[NB] dosya verisi
func buildFileContentPacket(data []byte) []byte {
	pktLen := tcpHeaderLength + len(data)
	pkt := make([]byte, pktLen)
	binary.LittleEndian.PutUint16(pkt[0:2], uint16(pktLen))
	binary.LittleEndian.PutUint16(pkt[2:4], uint16(CmdFileContentAsk))
	copy(pkt[tcpHeaderLength:], data)
	return pkt
}

// buildFileEndPacket, dosya transfer bitiş paketi oluşturur.
//
// Paket Formatı (toplam 4 byte):
//
//	[2B] length = 4
//	[2B] cmd = 0x8005 (CmdFileEndAsk)
func buildFileEndPacket() []byte {
	pkt := make([]byte, 4)
	binary.LittleEndian.PutUint16(pkt[0:2], 4)
	binary.LittleEndian.PutUint16(pkt[2:4], uint16(CmdFileEndAsk))
	return pkt
}

// parseFileStartResponse, CmdFileStartAnswer paketini ayrıştırır.
//
// Dönen değerler:
//   - errCode: İşlem sonuç kodu
//   - existBytes: Daha önce gönderilmiş byte sayısı (resume desteği)
//   - ok: Ayrıştırma başarılı mı
func parseFileStartResponse(data []byte) (errCode ErrorCode, existBytes uint32, ok bool) {
	if len(data) < 10 {
		return 0, 0, false
	}
	errCode = ErrorCode(binary.LittleEndian.Uint16(data[4:6]))
	existBytes = binary.LittleEndian.Uint32(data[6:10])
	return errCode, existBytes, true
}

// parseFileEndResponse, CmdFileEndAnswer paketini ayrıştırır.
func parseFileEndResponse(data []byte) (errCode ErrorCode, ok bool) {
	if len(data) < 6 {
		return 0, false
	}
	errCode = ErrorCode(binary.LittleEndian.Uint16(data[4:6]))
	return errCode, true
}

// buildUDPScanPacket, ağda cihaz arama için UDP broadcast paketi oluşturur.
// UDP port 10001'e broadcast olarak gönderilir.
//
// Paket Formatı:
//
//	[2B] length
//	[2B] cmd = 0x1001 (CmdSearchDeviceAsk)
//	[4B] versiyon = transportVersion
func buildUDPScanPacket() []byte {
	pkt := make([]byte, 8)
	binary.LittleEndian.PutUint16(pkt[0:2], 8)
	binary.LittleEndian.PutUint16(pkt[2:4], uint16(CmdSearchDeviceAsk))
	binary.LittleEndian.PutUint32(pkt[4:8], transportVersion)
	return pkt
}
