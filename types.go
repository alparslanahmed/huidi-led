package huidu

import (
	"fmt"
	"io"
	"time"
)

// ─── Protokol Sabitleri ─────────────────────────────────────────────────────────

const (
	// DefaultPort, Huidu cihazlarının varsayılan TCP dinleme portudur.
	DefaultPort = 10001

	// DefaultTimeout, TCP işlemleri için varsayılan zaman aşımı süresidir.
	DefaultTimeout = 10 * time.Second

	// DefaultHeartbeatInterval, heartbeat paketlerinin gönderilme aralığıdır.
	// SDK orijinal kodunda 30 saniye olarak belirlenmiştir.
	DefaultHeartbeatInterval = 30 * time.Second

	// MaxContentLength, tek bir TCP paketinde taşınabilecek maksimum veri boyutudur.
	// SDK'daki _maxContentLength sabitiyle aynıdır.
	MaxContentLength = 8000

	// tcpHeaderLength, temel TCP paket başlık uzunluğudur.
	// Format: [2B length LE][2B command LE]
	tcpHeaderLength = 4

	// sdkCmdHeaderLength, SDK komut paketlerinin başlık uzunluğudur.
	// Format: [2B length][2B cmd][4B total XML length][4B XML offset]
	sdkCmdHeaderLength = 12

	// transportVersion, transport protokol versiyon numarasıdır.
	// SDK'daki _LOCAL_TCP_VERSION sabitiyle aynıdır.
	transportVersion uint32 = 0x1000005

	// sdkVersion, SDK protokol versiyon numarasıdır.
	// SDK'daki _SDK_VERSION sabitiyle aynıdır.
	sdkVersion uint32 = 0x1000000

	// maxDeviceIDLength, cihaz ID'sinin maksimum uzunluğudur.
	maxDeviceIDLength = 15
)

// ─── Komut Tipleri ──────────────────────────────────────────────────────────────

// CmdType, Huidu binary protokolündeki komut tiplerini temsil eder.
// Her TCP paketinin 3. ve 4. byte'ları (little-endian) komut tipini belirtir.
type CmdType uint16

const (
	// CmdHeartbeatAsk, TCP heartbeat isteğidir. Bağlantıyı canlı tutmak için
	// periyodik olarak gönderilir. 4 byte'lık minimal pakettir.
	CmdHeartbeatAsk CmdType = 0x005f

	// CmdHeartbeatAnswer, heartbeat isteğine cihazın verdiği yanıttır.
	CmdHeartbeatAnswer CmdType = 0x0060

	// CmdSearchDeviceAsk, ağda cihaz aramak için UDP broadcast olarak gönderilir.
	CmdSearchDeviceAsk CmdType = 0x1001

	// CmdSearchDeviceAnswer, cihaz arama isteğine verilen yanıttır.
	// Cihaz bilgilerini (ID, IP vb.) içerir.
	CmdSearchDeviceAnswer CmdType = 0x1002

	// CmdErrorAnswer, herhangi bir komuta hata yanıtıdır.
	// Veri bölümünde 2 byte'lık hata kodu bulunur.
	CmdErrorAnswer CmdType = 0x2000

	// CmdServiceAsk, transport protokol versiyon anlaşma isteğidir.
	// Bağlantının ilk adımı olarak gönderilir.
	// Format: [2B len=8][2B cmd=0x2001][4B version]
	CmdServiceAsk CmdType = 0x2001

	// CmdServiceAnswer, versiyon anlaşma yanıtıdır.
	// Cihazın desteklediği versiyon numarasını içerir.
	CmdServiceAnswer CmdType = 0x2002

	// CmdSdkCmdAsk, XML tabanlı SDK komut isteğidir.
	// 12 byte header + XML verisi formatındadır.
	// Header: [2B len][2B cmd=0x2003][4B total XML len][4B XML offset]
	CmdSdkCmdAsk CmdType = 0x2003

	// CmdSdkCmdAnswer, SDK komut yanıtıdır.
	// Aynı 12 byte header formatını kullanır.
	CmdSdkCmdAnswer CmdType = 0x2004

	// CmdGPSInfoAnswer, GPS bilgi yanıtıdır.
	CmdGPSInfoAnswer CmdType = 0x3007

	// CmdFileStartAsk, dosya transfer başlatma isteğidir.
	// Dosya adı, MD5 hash, boyut ve tip bilgilerini içerir.
	CmdFileStartAsk CmdType = 0x8001

	// CmdFileStartAnswer, dosya transfer başlatma yanıtıdır.
	// Hata kodu ve daha önce gönderilmiş byte sayısını içerir (resume desteği).
	CmdFileStartAnswer CmdType = 0x8002

	// CmdFileContentAsk, dosya içeriği taşıyan pakettir.
	// [4B header][dosya verisi] formatındadır.
	CmdFileContentAsk CmdType = 0x8003

	// CmdFileContentAnswer, dosya içeriği alındı yanıtıdır.
	CmdFileContentAnswer CmdType = 0x8004

	// CmdFileEndAsk, dosya transferi bitiş isteğidir.
	CmdFileEndAsk CmdType = 0x8005

	// CmdFileEndAnswer, dosya transferi bitiş yanıtıdır.
	CmdFileEndAnswer CmdType = 0x8006

	// CmdReadFileAsk, cihazdan dosya okuma isteğidir.
	CmdReadFileAsk CmdType = 0x8007

	// CmdReadFileAnswer, cihazdan dosya okuma yanıtıdır.
	CmdReadFileAnswer CmdType = 0x8008
)

// String, CmdType'ın okunabilir string temsilini döner.
func (c CmdType) String() string {
	switch c {
	case CmdHeartbeatAsk:
		return "HeartbeatAsk"
	case CmdHeartbeatAnswer:
		return "HeartbeatAnswer"
	case CmdSearchDeviceAsk:
		return "SearchDeviceAsk"
	case CmdSearchDeviceAnswer:
		return "SearchDeviceAnswer"
	case CmdErrorAnswer:
		return "ErrorAnswer"
	case CmdServiceAsk:
		return "ServiceAsk"
	case CmdServiceAnswer:
		return "ServiceAnswer"
	case CmdSdkCmdAsk:
		return "SdkCmdAsk"
	case CmdSdkCmdAnswer:
		return "SdkCmdAnswer"
	case CmdFileStartAsk:
		return "FileStartAsk"
	case CmdFileStartAnswer:
		return "FileStartAnswer"
	case CmdFileContentAsk:
		return "FileContentAsk"
	case CmdFileContentAnswer:
		return "FileContentAnswer"
	case CmdFileEndAsk:
		return "FileEndAsk"
	case CmdFileEndAnswer:
		return "FileEndAnswer"
	case CmdReadFileAsk:
		return "ReadFileAsk"
	case CmdReadFileAnswer:
		return "ReadFileAnswer"
	default:
		return fmt.Sprintf("Unknown(0x%04x)", uint16(c))
	}
}

// ─── Hata Kodları ───────────────────────────────────────────────────────────────

// ErrorCode, Huidu cihazından dönen hata kodlarını temsil eder.
type ErrorCode int

const (
	ErrSuccess           ErrorCode = 0  // Başarılı
	ErrWriteFinish       ErrorCode = 1  // Dosya yazma tamamlandı
	ErrProcessError      ErrorCode = 2  // İşlem akış hatası
	ErrVersionTooLow     ErrorCode = 3  // Protokol versiyonu çok düşük
	ErrDeviceOccupied    ErrorCode = 4  // Cihaz başka bir istemci tarafından kullanılıyor
	ErrFileOccupied      ErrorCode = 5  // Dosya kullanımda
	ErrReadFileExcessive ErrorCode = 6  // Çok fazla dosya okuma isteği
	ErrInvalidPacketLen  ErrorCode = 7  // Paket uzunluğu hatalı
	ErrInvalidParam      ErrorCode = 8  // Geçersiz parametre
	ErrNotSpaceToSave    ErrorCode = 9  // Yetersiz depolama alanı
	ErrCreateFileFailed  ErrorCode = 10 // Dosya oluşturma hatası
	ErrWriteFileFailed   ErrorCode = 11 // Dosya yazma hatası
	ErrReadFileFailed    ErrorCode = 12 // Dosya okuma hatası
	ErrInvalidFileData   ErrorCode = 13 // Geçersiz dosya verisi
	ErrFileContentError  ErrorCode = 14 // Dosya içeriği hatalı
	ErrOpenFileFailed    ErrorCode = 15 // Dosya açma hatası
	ErrSeekFileFailed    ErrorCode = 16 // Dosya konum hatası
	ErrRenameFailed      ErrorCode = 17 // Yeniden adlandırma hatası
	ErrFileNotFound      ErrorCode = 18 // Dosya bulunamadı
	ErrFileNotFinish     ErrorCode = 19 // Dosya alımı tamamlanmadı
	ErrXmlCmdTooLong     ErrorCode = 20 // XML komutu çok uzun
	ErrInvalidXmlIndex   ErrorCode = 21 // Geçersiz XML index değeri
	ErrParseXmlFailed    ErrorCode = 22 // XML ayrıştırma hatası
	ErrInvalidMethod     ErrorCode = 23 // Geçersiz metot adı
	ErrMemoryFailed      ErrorCode = 24 // Bellek hatası
	ErrSystemError       ErrorCode = 25 // Sistem hatası
	ErrUnsupportVideo    ErrorCode = 26 // Desteklenmeyen video
	ErrNotMediaFile      ErrorCode = 27 // Medya dosyası değil
	ErrParseVideoFailed  ErrorCode = 28 // Video ayrıştırma hatası
	ErrUnsupportFPS      ErrorCode = 29 // Desteklenmeyen kare hızı
	ErrUnsupportRes      ErrorCode = 30 // Desteklenmeyen çözünürlük
	ErrUnsupportFormat   ErrorCode = 31 // Desteklenmeyen format
	ErrUnsupportDuration ErrorCode = 32 // Desteklenmeyen süre
	ErrDownloadFailed    ErrorCode = 33 // Dosya indirme hatası
	ErrScreenNodeNull    ErrorCode = 34 // Ekran düğümü bulunamadı
	ErrNodeExist         ErrorCode = 35 // Düğüm zaten mevcut
	ErrNodeNotExist      ErrorCode = 36 // Düğüm mevcut değil
	ErrPluginNotExist    ErrorCode = 37 // Plugin mevcut değil
	ErrCheckLicense      ErrorCode = 38 // Lisans doğrulama hatası
	ErrNotFoundWifi      ErrorCode = 39 // WiFi modülü bulunamadı
	ErrTestWifiFailed    ErrorCode = 40 // WiFi testi başarısız
	ErrRunningError      ErrorCode = 41 // Çalışma hatası
	ErrUnsupportMethod   ErrorCode = 42 // Desteklenmeyen metot
	ErrInvalidGUID       ErrorCode = 43 // Geçersiz GUID
	ErrFirmwareFormat    ErrorCode = 44 // Firmware format hatası
	ErrTagNotFound       ErrorCode = 45 // Etiket bulunamadı
	ErrAttrNotFound      ErrorCode = 46 // Özellik bulunamadı
	ErrCreateTagFailed   ErrorCode = 47 // Etiket oluşturma hatası
	ErrUnsupportDevice   ErrorCode = 48 // Desteklenmeyen cihaz modeli
	ErrPermissionDenied  ErrorCode = 49 // Yetersiz izin
	ErrPasswdTooSimple   ErrorCode = 50 // Şifre çok basit
)

// String, ErrorCode'un okunabilir açıklamasını döner.
func (e ErrorCode) String() string {
	names := map[ErrorCode]string{
		ErrSuccess:          "Başarılı",
		ErrWriteFinish:      "Dosya yazma tamamlandı",
		ErrProcessError:     "İşlem akış hatası",
		ErrVersionTooLow:    "Protokol versiyonu çok düşük",
		ErrDeviceOccupied:   "Cihaz meşgul",
		ErrFileOccupied:     "Dosya kullanımda",
		ErrInvalidPacketLen: "Paket uzunluğu hatalı",
		ErrInvalidParam:     "Geçersiz parametre",
		ErrNotSpaceToSave:   "Yetersiz depolama alanı",
		ErrCreateFileFailed: "Dosya oluşturma hatası",
		ErrWriteFileFailed:  "Dosya yazma hatası",
		ErrReadFileFailed:   "Dosya okuma hatası",
		ErrParseXmlFailed:   "XML ayrıştırma hatası",
		ErrInvalidMethod:    "Geçersiz metot adı",
		ErrMemoryFailed:     "Bellek hatası",
		ErrSystemError:      "Sistem hatası",
		ErrUnsupportVideo:   "Desteklenmeyen video",
		ErrUnsupportMethod:  "Desteklenmeyen metot",
		ErrInvalidGUID:      "Geçersiz GUID",
		ErrPermissionDenied: "Yetersiz izin",
		ErrFileNotFound:     "Dosya bulunamadı",
		ErrUnsupportDevice:  "Desteklenmeyen cihaz",
		ErrNotFoundWifi:     "WiFi modülü bulunamadı",
	}
	if name, ok := names[e]; ok {
		return name
	}
	return fmt.Sprintf("Bilinmeyen hata (%d)", int(e))
}

// Error, ErrorCode'u error interface'i olarak kullanılabilir hale getirir.
func (e ErrorCode) Error() string {
	return fmt.Sprintf("huidu error %d: %s", int(e), e.String())
}

// ─── SDK Metot Adları ───────────────────────────────────────────────────────────

// SdkMethod, SDK XML komutlarında kullanılan metot adlarını tanımlar.
// Her metot, cihaza gönderilen XML'deki "method" attribute'una karşılık gelir.
type SdkMethod string

const (
	// MethodGetIFVersion, SDK protokol versiyonunu sorgular.
	// Bağlantı kurulumunda otomatik olarak çağrılır.
	MethodGetIFVersion SdkMethod = "GetIFVersion"

	// MethodAddProgram, yeni bir program (ekran içeriği) gönderir.
	// Mevcut tüm programları değiştirir.
	MethodAddProgram SdkMethod = "AddProgram"

	// MethodUpdateProgram, mevcut bir programı günceller.
	MethodUpdateProgram SdkMethod = "UpdateProgram"

	// MethodDeleteProgram, belirtilen programı siler.
	// İç içerik boş bırakılırsa tüm programlar silinir.
	MethodDeleteProgram SdkMethod = "DeleteProgram"

	// MethodGetAllFontInfo, cihazda yüklü font bilgilerini sorgular.
	MethodGetAllFontInfo SdkMethod = "GetAllFontInfo"

	// MethodGetLuminancePloy, parlaklık ayar bilgilerini sorgular.
	MethodGetLuminancePloy SdkMethod = "GetLuminancePloy"

	// MethodSetLuminancePloy, parlaklık ayarlarını gerçekleştirir.
	MethodSetLuminancePloy SdkMethod = "SetLuminancePloy"

	// MethodGetSwitchTime, zamanlı açma/kapama bilgilerini sorgular.
	MethodGetSwitchTime SdkMethod = "GetSwitchTime"

	// MethodSetSwitchTime, zamanlı açma/kapama ayarlarını yapar.
	MethodSetSwitchTime SdkMethod = "SetSwitchTime"

	// MethodOpenScreen, ekranı hemen açar.
	MethodOpenScreen SdkMethod = "OpenScreen"

	// MethodCloseScreen, ekranı hemen kapatır.
	MethodCloseScreen SdkMethod = "CloseScreen"

	// MethodGetTimeInfo, cihaz zaman bilgilerini sorgular.
	MethodGetTimeInfo SdkMethod = "GetTimeInfo"

	// MethodSetTimeInfo, cihaz zamanını ayarlar.
	MethodSetTimeInfo SdkMethod = "SetTimeInfo"

	// MethodGetBootLogo, açılış logosu bilgisini sorgular.
	MethodGetBootLogo SdkMethod = "GetBootLogo"

	// MethodSetBootLogoName, açılış logosunu ayarlar.
	MethodSetBootLogoName SdkMethod = "SetBootLogoName"

	// MethodClearBootLogo, açılış logosunu temizler.
	MethodClearBootLogo SdkMethod = "ClearBootLogo"

	// MethodGetSDKTcpServer, cihazın bağlandığı TCP sunucu bilgisini sorgular.
	MethodGetSDKTcpServer SdkMethod = "GetSDKTcpServer"

	// MethodSetSDKTcpServer, cihazın bağlanacağı TCP sunucu bilgisini ayarlar.
	MethodSetSDKTcpServer SdkMethod = "SetSDKTcpServer"

	// MethodGetDeviceInfo, cihaz donanım ve yazılım bilgilerini sorgular.
	MethodGetDeviceInfo SdkMethod = "GetDeviceInfo"

	// MethodGetEth0Info, Ethernet ağ ayarlarını sorgular.
	MethodGetEth0Info SdkMethod = "GetEth0Info"

	// MethodSetEth0Info, Ethernet ağ ayarlarını yapar.
	MethodSetEth0Info SdkMethod = "SetEth0Info"

	// MethodGetPppoeInfo, 3G/4G bilgilerini sorgular.
	MethodGetPppoeInfo SdkMethod = "GetPppoeInfo"

	// MethodSetApn, APN bilgisini ayarlar.
	MethodSetApn SdkMethod = "SetApn"

	// MethodGetWifiInfo, WiFi bilgilerini sorgular.
	MethodGetWifiInfo SdkMethod = "GetWifiInfo"

	// MethodSetWifiInfo, WiFi ayarlarını yapar.
	MethodSetWifiInfo SdkMethod = "SetWifiInfo"

	// MethodGetFiles, cihaza yüklenmiş dosya listesini sorgular.
	MethodGetFiles SdkMethod = "GetFiles"

	// MethodDeleteFiles, belirtilen dosyaları siler.
	MethodDeleteFiles SdkMethod = "DeleteFiles"

	// MethodGetGpsRespondEnable, GPS bilgi raporlama durumunu sorgular.
	MethodGetGpsRespondEnable SdkMethod = "GetGpsRespondEnable"

	// MethodSetGpsRespondEnable, GPS bilgi raporlama durumunu ayarlar.
	MethodSetGpsRespondEnable SdkMethod = "SetGpsRespondEnable"

	// MethodGetMulScreenSync, çoklu ekran senkronizasyon durumunu sorgular.
	MethodGetMulScreenSync SdkMethod = "GetMulScreenSync"

	// MethodSetMulScreenSync, çoklu ekran senkronizasyonunu ayarlar.
	MethodSetMulScreenSync SdkMethod = "SetMulScreenSync"

	// MethodGetProgram, mevcut program bilgisini okur.
	MethodGetProgram SdkMethod = "GetProgram"

	// MethodGetCurrentPlayProgramGUID, şu an oynatılan programın GUID'ini döner.
	MethodGetCurrentPlayProgramGUID SdkMethod = "GetCurrentPlayProgramGUID"

	// MethodSetPlayTypeToNormal, normal oynatma moduna döner.
	MethodSetPlayTypeToNormal SdkMethod = "SetPlayTypeToNormal"
)

// ─── Efekt Tipleri ──────────────────────────────────────────────────────────────

// EffectType, metin ve görsel öğeleri için geçiş efekti tiplerini tanımlar.
// LED ekranda bir öğe gösterilirken veya değiştirilirken bu efektler kullanılır.
type EffectType int

const (
	EffectImmediate       EffectType = 0  // Anında göster (geçiş efekti yok)
	EffectLeftMove        EffectType = 1  // Sola kayma
	EffectRightMove       EffectType = 2  // Sağa kayma
	EffectUpMove          EffectType = 3  // Yukarı kayma
	EffectDownMove        EffectType = 4  // Aşağı kayma
	EffectLeftCover       EffectType = 5  // Soldan örtme
	EffectRightCover      EffectType = 6  // Sağdan örtme
	EffectUpCover         EffectType = 7  // Yukarıdan örtme
	EffectDownCover       EffectType = 8  // Aşağıdan örtme
	EffectLeftTopCover    EffectType = 9  // Sol üstten örtme
	EffectLeftBotCover    EffectType = 10 // Sol alttan örtme
	EffectRightTopCover   EffectType = 11 // Sağ üstten örtme
	EffectRightBotCover   EffectType = 12 // Sağ alttan örtme
	EffectHorizDivide     EffectType = 13 // Yatay bölme
	EffectVertDivide      EffectType = 14 // Dikey bölme
	EffectHorizClose      EffectType = 15 // Yatay kapanma
	EffectVertClose       EffectType = 16 // Dikey kapanma
	EffectFade            EffectType = 17 // Solma (fade in/out)
	EffectHorizShutter    EffectType = 18 // Yatay jaluzi
	EffectVertShutter     EffectType = 19 // Dikey jaluzi
	EffectNoClear         EffectType = 20 // Ekran temizlemeden göster
	EffectLeftScroll      EffectType = 21 // Sürekli sola kaydırma
	EffectRightScroll     EffectType = 22 // Sürekli sağa kaydırma
	EffectUpScroll        EffectType = 23 // Sürekli yukarı kaydırma
	EffectDownScroll      EffectType = 24 // Sürekli aşağı kaydırma
	EffectRandom          EffectType = 25 // Rastgele efekt
	EffectLeftScrollLoop  EffectType = 26 // Baştan sona bağlı sürekli sola kaydırma
	EffectRightScrollLoop EffectType = 27 // Baştan sona bağlı sürekli sağa kaydırma
	EffectUpScrollLoop    EffectType = 28 // Baştan sona bağlı sürekli yukarı kaydırma
	EffectDownScrollLoop  EffectType = 29 // Baştan sona bağlı sürekli aşağı kaydırma
)

// String, EffectType'ın okunabilir adını döner.
func (e EffectType) String() string {
	names := map[EffectType]string{
		EffectImmediate:       "Anında",
		EffectLeftMove:        "Sola Kayma",
		EffectRightMove:       "Sağa Kayma",
		EffectUpMove:          "Yukarı Kayma",
		EffectDownMove:        "Aşağı Kayma",
		EffectFade:            "Solma",
		EffectRandom:          "Rastgele",
		EffectLeftScroll:      "Sürekli Sola",
		EffectRightScroll:     "Sürekli Sağa",
		EffectUpScroll:        "Sürekli Yukarı",
		EffectDownScroll:      "Sürekli Aşağı",
		EffectLeftScrollLoop:  "Döngülü Sola",
		EffectRightScrollLoop: "Döngülü Sağa",
		EffectUpScrollLoop:    "Döngülü Yukarı",
		EffectDownScrollLoop:  "Döngülü Aşağı",
	}
	if name, ok := names[e]; ok {
		return name
	}
	return fmt.Sprintf("Efekt(%d)", int(e))
}

// IsContinuousScroll, efektin sürekli kayan türde olup olmadığını kontrol eder.
// Bu tür efektlerde metin singleLine=true olarak işaretlenmelidir.
func (e EffectType) IsContinuousScroll() bool {
	return e == EffectLeftScroll || e == EffectRightScroll ||
		e == EffectLeftScrollLoop || e == EffectRightScrollLoop
}

// IsVerticalScroll, efektin dikey kayan türde olup olmadığını kontrol eder.
func (e EffectType) IsVerticalScroll() bool {
	return e == EffectUpScroll || e == EffectDownScroll ||
		e == EffectUpScrollLoop || e == EffectDownScrollLoop
}

// ─── Dosya Tipleri ──────────────────────────────────────────────────────────────

// FileType, cihaza yüklenecek dosyanın tipini belirtir.
type FileType int

const (
	FileTypeAuto          FileType = -1  // Otomatik tespit
	FileTypeImage         FileType = 0   // Görsel dosyası (BMP, JPG, PNG, GIF vb.)
	FileTypeVideo         FileType = 1   // Video dosyası (MP4, AVI, MKV vb.)
	FileTypeFont          FileType = 2   // Font dosyası (TTF, TTC, BDF)
	FileTypeFirmware      FileType = 3   // Firmware güncelleme dosyası (BIN)
	FileTypeFPGAConfig    FileType = 4   // FPGA yapılandırma dosyası
	FileTypeSettingConfig FileType = 5   // Ayar yapılandırma dosyası
	FileTypeProgramXML    FileType = 9   // Program şablon XML dosyası
	FileTypeTempImage     FileType = 128 // Geçici görsel (toplam ≤ 10MB)
	FileTypeTempVideo     FileType = 129 // Geçici video (toplam ≤ 10MB)
)

// ─── Veri Yapıları ──────────────────────────────────────────────────────────────

// DeviceInfo, cihazın donanım ve yazılım bilgilerini tutar.
// GetDeviceInfo komutuyla alınır.
type DeviceInfo struct {
	CPU            string // İşlemci tipi (ör: "Freescale.iMax6", "TI.335x")
	Model          string // Kart modeli
	DeviceID       string // Benzersiz cihaz kimliği
	DeviceName     string // Cihaz adı
	FPGAVersion    string // FPGA versiyonu
	AppVersion     string // Firmware versiyonu
	KernelVersion  string // Linux kernel versiyonu
	ScreenWidth    int    // Ekran genişliği (piksel)
	ScreenHeight   int    // Ekran yüksekliği (piksel)
	ScreenRotation int    // Ekran dönme açısı (0, 90, 180, 270)
}

// EthernetInfo, Ethernet ağ arayüzü bilgilerini tutar.
type EthernetInfo struct {
	Enabled  bool   // Ethernet aktif mi
	AutoDHCP bool   // DHCP otomatik mi
	IP       string // IP adresi
	Netmask  string // Alt ağ maskesi
	Gateway  string // Ağ geçidi
	DNS      string // DNS sunucusu
}

// WifiAPInfo, bir WiFi erişim noktasının bilgilerini tutar.
type WifiAPInfo struct {
	SSID       string // Ağ adı
	Password   string // Şifre
	MAC        string // MAC adresi
	AutoDHCP   bool   // DHCP otomatik mi
	IP         string // IP adresi
	Netmask    string // Alt ağ maskesi
	Gateway    string // Ağ geçidi
	Channel    string // Kanal numarası
	Encryption string // Şifreleme türü (WPA-PSK vb.)
}

// WifiInfo, WiFi modülünün tüm bilgilerini tutar.
type WifiInfo struct {
	HasWifi     bool       // WiFi modülü var mı
	Enabled     bool       // WiFi aktif mi
	WorkMode    int        // 0: AP modu, 1: Station modu
	APInfo      WifiAPInfo // AP modu bilgileri
	StationSSID string     // Station modu: bağlanılan ağ adı
	StationPass string     // Station modu: şifre
}

// ServerInfo, cihazın bağlandığı TCP sunucu bilgisini tutar.
type ServerInfo struct {
	Host string // Sunucu IP veya domain adı
	Port int    // Sunucu portu
}

// TimeInfo, cihaz zaman ayarlarını tutar.
type TimeInfo struct {
	Timezone string // Saat dilimi (ör: "(UTC+03:00)Istanbul")
	Summer   bool   // Yaz saati uygulaması aktif mi
	Sync     string // Senkronizasyon modu: "none", "gps", "network", "auto"
	Time     string // Zaman değeri: "YYYY-MM-DD hh:mm:ss" (sync="none" ise geçerli)
}

// LuminanceInfo, parlaklık ayar bilgilerini tutar.
type LuminanceInfo struct {
	Mode         int             // 0: varsayılan, 1: zamanlı, 2: sensör
	DefaultValue int             // Varsayılan parlaklık değeri (1-100)
	CustomItems  []LuminanceItem // Zamanlı parlaklık öğeleri
	SensorMin    int             // Sensör minimum değeri (1-100)
	SensorMax    int             // Sensör maksimum değeri (1-100)
	SensorTime   int             // Sensör geçiş süresi (5-15 sn)
}

// LuminanceItem, zamanlı parlaklık programındaki bir zaman dilimini temsil eder.
type LuminanceItem struct {
	Enabled bool   // Bu zaman dilimi aktif mi
	Start   string // Başlangıç zamanı (hh:mm:ss)
	Percent int    // Parlaklık yüzdesi (1-100)
}

// SwitchTimeInfo, zamanlı açma/kapama bilgilerini tutar.
type SwitchTimeInfo struct {
	OpenEnabled bool             // Ekran varsayılan olarak açık mı
	PloyEnabled bool             // Zamanlı kontrol aktif mi
	Items       []SwitchTimeItem // Zamanlama öğeleri
}

// SwitchTimeItem, bir zamanlı açma/kapama kuralını temsil eder.
type SwitchTimeItem struct {
	Enabled bool   // Bu kural aktif mi (true: aç, false: kapat)
	Start   string // Başlangıç zamanı (hh:mm:ss)
	End     string // Bitiş zamanı (hh:mm:ss)
}

// BootLogoInfo, açılış logosu bilgilerini tutar.
type BootLogoInfo struct {
	Exists bool   // Logo ayarlanmış mı
	Name   string // Logo dosyası adı
	MD5    string // Logo dosyasının MD5 hash'i
}

// FontInfo, cihazda yüklü bir fontu temsil eder.
type FontInfo struct {
	FontName  string // Font görünen adı
	FileName  string // Font dosya adı
	Bold      bool   // Kalın
	Italic    bool   // İtalik
	Underline bool   // Altı çizili
}

// FileInfo, cihazda yüklü bir dosyayı temsil eder.
type FileInfo struct {
	Name      string // Dosya adı
	Size      int64  // Dosya boyutu (byte)
	ExistSize int64  // Mevcut boyut (kısmi yükleme durumunda)
	MD5       string // MD5 hash
	Type      string // Dosya tipi
}

// UploadProgress, dosya yükleme ilerleme bilgisini taşır.
type UploadProgress struct {
	FileName   string  // Yüklenen dosya adı
	TotalBytes int64   // Toplam dosya boyutu
	SentBytes  int64   // Gönderilen byte sayısı
	Percent    float64 // İlerleme yüzdesi (0-100)
}

// ─── Seçenek Yapıları ───────────────────────────────────────────────────────────

// DeviceOption, Device yapılandırma seçeneklerini tanımlar.
// Functional Options pattern kullanılır.
type DeviceOption func(*deviceOptions)

type deviceOptions struct {
	timeout           time.Duration
	heartbeatInterval time.Duration
	autoReconnect     bool
	logger            Logger
	onProgress        func(UploadProgress)
}

func defaultDeviceOptions() deviceOptions {
	return deviceOptions{
		timeout:           DefaultTimeout,
		heartbeatInterval: DefaultHeartbeatInterval,
		autoReconnect:     false,
		logger:            nil,
		onProgress:        nil,
	}
}

// WithTimeout, TCP işlemleri için zaman aşımı süresini ayarlar.
//
//	device := huidu.NewDevice("192.168.6.1", 10001,
//	    huidu.WithTimeout(5 * time.Second),
//	)
func WithTimeout(d time.Duration) DeviceOption {
	return func(o *deviceOptions) {
		o.timeout = d
	}
}

// WithHeartbeatInterval, heartbeat paket gönderim aralığını ayarlar.
// Varsayılan değer 30 saniyedir. Daha kısa aralık bağlantı kararlılığını
// artırır ancak ağ trafiğini artırır.
func WithHeartbeatInterval(d time.Duration) DeviceOption {
	return func(o *deviceOptions) {
		o.heartbeatInterval = d
	}
}

// WithAutoReconnect, bağlantı koptuğunda otomatik yeniden bağlanmayı aktifleştirir.
func WithAutoReconnect(enabled bool) DeviceOption {
	return func(o *deviceOptions) {
		o.autoReconnect = enabled
	}
}

// WithLogger, özel bir loglama arayüzü ayarlar.
// Varsayılan olarak loglama devre dışıdır.
func WithLogger(l Logger) DeviceOption {
	return func(o *deviceOptions) {
		o.logger = l
	}
}

// WithProgressCallback, dosya yükleme ilerleme callback'i ayarlar.
func WithProgressCallback(fn func(UploadProgress)) DeviceOption {
	return func(o *deviceOptions) {
		o.onProgress = fn
	}
}

// ─── Logger Arayüzü ─────────────────────────────────────────────────────────────

// Logger, kütüphanenin loglama arayüzüdür.
// stdlib log paketi veya zerolog/zap gibi kütüphanelerle uyumludur.
type Logger interface {
	// Printf, formatlanmış bir log mesajı yazar.
	Printf(format string, v ...interface{})
}

// ─── Hizalama Tipleri ───────────────────────────────────────────────────────────

// HAlign, yatay hizalama tiplerini tanımlar.
type HAlign string

const (
	HAlignLeft   HAlign = "left"   // Sola hizalı
	HAlignCenter HAlign = "center" // Ortaya hizalı
	HAlignRight  HAlign = "right"  // Sağa hizalı
)

// VAlign, dikey hizalama tiplerini tanımlar.
type VAlign string

const (
	VAlignTop    VAlign = "top"    // Üste hizalı
	VAlignMiddle VAlign = "middle" // Ortaya hizalı
	VAlignBottom VAlign = "bottom" // Alta hizalı
)

// ─── Saat Tipleri ───────────────────────────────────────────────────────────────

// ClockType, saat görüntüleme tipini tanımlar.
type ClockType string

const (
	ClockDigital ClockType = "digital" // Dijital saat
	ClockDial    ClockType = "dial"    // Analog saat (kadran)
)

// ─── Görsel Yerleştirme ─────────────────────────────────────────────────────────

// ImageFit, görselin alana nasıl yerleştirileceğini belirler.
type ImageFit string

const (
	ImageFitFill    ImageFit = "fill"    // Alanı doldur
	ImageFitCenter  ImageFit = "center"  // Ortala
	ImageFitStretch ImageFit = "stretch" // Uzat
	ImageFitTile    ImageFit = "tile"    // Döşe
)

// ─── Program Tipleri ────────────────────────────────────────────────────────────

// ProgramType, program tipini belirler.
type ProgramType string

const (
	ProgramNormal   ProgramType = "normal"   // Normal program
	ProgramTemplate ProgramType = "template" // Şablon program
	ProgramHTML5    ProgramType = "html5"    // HTML5 program
	ProgramOffline  ProgramType = "offline"  // Çevrimdışı program
)

// ─── Reader arayüzü ─────────────────────────────────────────────────────────────

// Kullanılmayan import kontrolü
var _ = io.EOF
