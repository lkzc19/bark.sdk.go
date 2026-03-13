package bark

import (
	"bytes"
	"encoding/base64"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// before 加载测试环境配置
func before(t *testing.T) string {
	t.Helper()
	if err := godotenv.Load(); err != nil {
		t.Skip("未找到 .env 文件，跳过集成测试")
	}
	key := os.Getenv("DeviceKey")
	if key == "" {
		t.Skip("环境变量 DeviceKey 未配置，跳过集成测试")
	}
	return key
}

// --- 参数校验单元测试（无需网络） ---

func TestValidate_MissingDeviceKey(t *testing.T) {
	err := validateReq(Req{Body: "内容"})
	if err == nil {
		t.Fatal("预期返回错误：DeviceKey 为空")
	}
}

func TestValidate_MissingTitleAndBody(t *testing.T) {
	err := validateReq(Req{DeviceKey: "key"})
	if err == nil {
		t.Fatal("预期返回错误：Title 和 Body 均为空")
	}
}

func TestValidate_InvalidLevel(t *testing.T) {
	err := validateReq(Req{DeviceKey: "key", Body: "内容", Level: "invalid"})
	if err == nil {
		t.Fatal("预期返回错误：Level 值非法")
	}
}

func TestValidate_ValidLevels(t *testing.T) {
	levels := []Level{Active, TimeSensitive, Passive, Critical}
	for _, lv := range levels {
		err := validateReq(Req{DeviceKey: "key", Body: "内容", Level: lv})
		if err != nil {
			t.Fatalf("Level=%s 不应返回错误：%v", lv, err)
		}
	}
}

// --- buildPostBody 单元测试 ---

func TestBuildPostBody_Call(t *testing.T) {
	b := buildPostBody(Req{Call: true})
	if b.Call != 1 {
		t.Fatalf("期望 Call=1，实际 Call=%d", b.Call)
	}
}

func TestBuildPostBody_AutoCopy(t *testing.T) {
	b := buildPostBody(Req{AutoCopy: true})
	if b.AutoCopy != 1 {
		t.Fatalf("期望 AutoCopy=1，实际 AutoCopy=%d", b.AutoCopy)
	}
}

func TestBuildPostBody_IsArchive_False(t *testing.T) {
	f := false
	b := buildPostBody(Req{IsArchive: &f})
	if b.IsArchive == nil || *b.IsArchive != 0 {
		t.Fatal("期望 isArchive=0")
	}
}

func TestBuildPostBody_IsArchive_True(t *testing.T) {
	tr := true
	b := buildPostBody(Req{IsArchive: &tr})
	if b.IsArchive == nil || *b.IsArchive != 1 {
		t.Fatal("期望 isArchive=1")
	}
}

func TestBuildPostBody_IsArchive_Nil(t *testing.T) {
	b := buildPostBody(Req{})
	if b.IsArchive != nil {
		t.Fatal("IsArchive 为 nil 时，postBody.IsArchive 也应为 nil（使用 API 默认值）")
	}
}

func TestBuildPostBody_Critical_DefaultVolume(t *testing.T) {
	b := buildPostBody(Req{Level: Critical})
	if b.Volume == nil || *b.Volume != 5 {
		t.Fatalf("Critical 未指定 Volume 时，期望默认值 5，实际 %v", b.Volume)
	}
}

func TestBuildPostBody_Critical_CustomVolume(t *testing.T) {
	v := 8
	b := buildPostBody(Req{Level: Critical, Volume: &v})
	if b.Volume == nil || *b.Volume != 8 {
		t.Fatalf("期望 Volume=8，实际 %v", b.Volume)
	}
}

func TestBuildPostBody_Critical_VolumeClamp(t *testing.T) {
	over := 100
	b := buildPostBody(Req{Level: Critical, Volume: &over})
	if b.Volume == nil || *b.Volume != 10 {
		t.Fatalf("Volume 超过 10 时应截断为 10，实际 %v", b.Volume)
	}

	under := -5
	b = buildPostBody(Req{Level: Critical, Volume: &under})
	if b.Volume == nil || *b.Volume != 0 {
		t.Fatalf("Volume 低于 0 时应截断为 0，实际 %v", b.Volume)
	}
}

func TestBuildPostBody_NonCritical_NoVolume(t *testing.T) {
	b := buildPostBody(Req{Level: Active})
	if b.Volume != nil {
		t.Fatal("非 Critical 级别不应设置 Volume")
	}
}

// --- 加密校验单元测试 ---

func TestValidate_Encrypt_InvalidAlgorithm(t *testing.T) {
	err := validateEncryptConfig(&EncryptConfig{Algorithm: "AES512", Key: "1234567890123456"})
	if err == nil {
		t.Fatal("非法 Algorithm 应返回错误")
	}
}

func TestValidate_Encrypt_InvalidMode(t *testing.T) {
	err := validateEncryptConfig(&EncryptConfig{Key: "1234567890123456", Mode: "CFB"})
	if err == nil {
		t.Fatal("非法 Mode 应返回错误")
	}
}

func TestValidate_Encrypt_KeyLen_AES128(t *testing.T) {
	err := validateEncryptConfig(&EncryptConfig{Algorithm: AES128, Key: "short"})
	if err == nil {
		t.Fatal("AES128 密钥不足 16 字节应报错")
	}
}

func TestValidate_Encrypt_KeyLen_AES192(t *testing.T) {
	err := validateEncryptConfig(&EncryptConfig{Algorithm: AES192, Key: "1234567890123456"}) // 16字节，不够
	if err == nil {
		t.Fatal("AES192 密钥不足 24 字节应报错")
	}
}

func TestValidate_Encrypt_KeyLen_AES256(t *testing.T) {
	err := validateEncryptConfig(&EncryptConfig{Algorithm: AES256, Key: "12345678901234561234567890123456"}) // 32字节
	if err != nil {
		t.Fatalf("AES256 32字节密钥不应报错: %v", err)
	}
}

func TestValidate_Encrypt_CBC_IVLength(t *testing.T) {
	// CBC IV 不为 16 字节应报错
	err := validateEncryptConfig(&EncryptConfig{Algorithm: AES128, Mode: CBC, Key: "1234567890123456", IV: "short"})
	if err == nil {
		t.Fatal("CBC IV 不足 16 字节应返回错误")
	}
}

func TestValidate_Encrypt_GCM_NonceLength(t *testing.T) {
	// GCM Nonce 不为 12 字节应报错
	err := validateEncryptConfig(&EncryptConfig{Algorithm: AES128, Mode: GCM, Key: "1234567890123456", IV: "tooshort"})
	if err == nil {
		t.Fatal("GCM Nonce 不足 12 字节应返回错误")
	}
}

func TestValidate_Encrypt_ECB_IgnoresIV(t *testing.T) {
	// ECB 模式 IV 字段被忽略，不报错
	err := validateEncryptConfig(&EncryptConfig{Algorithm: AES128, Mode: ECB, Key: "1234567890123456", IV: "any_value_ignored"})
	if err != nil {
		t.Fatalf("ECB 模式 IV 应被忽略，不应报错: %v", err)
	}
}

func TestValidate_Encrypt_DefaultAlgoAndMode(t *testing.T) {
	// Algorithm 和 Mode 为空时使用默认值 AES128+CBC，不报错
	err := validateEncryptConfig(&EncryptConfig{Key: "1234567890123456"})
	if err != nil {
		t.Fatalf("默认 AES128+CBC 不应报错: %v", err)
	}
}

// --- PKCS7 填充测试 ---

func TestPkcs7Pad_FullBlock(t *testing.T) {
	// 16 字节数据（恰好一块）填充后应为 32 字节
	data := bytes.Repeat([]byte("a"), 16)
	padded := pkcs7Pad(data, 16)
	if len(padded) != 32 {
		t.Fatalf("期望填充后长度 32，实际 %d", len(padded))
	}
	for _, b := range padded[16:] {
		if b != 16 {
			t.Fatalf("期望填充字节值 16，实际 %d", b)
		}
	}
}

func TestPkcs7Pad_PartialBlock(t *testing.T) {
	// 10 字节数据，需要填充 6 字节
	data := bytes.Repeat([]byte("a"), 10)
	padded := pkcs7Pad(data, 16)
	if len(padded) != 16 {
		t.Fatalf("期望填充后长度 16，实际 %d", len(padded))
	}
}

// --- 各算法+模式加密单元测试 ---

// encryptCases 测试用例表，覆盖所有合法的 Algorithm × Mode 组合
var encryptCases = []struct {
	name string
	cfg  EncryptConfig
}{
	{name: "AES128_CBC", cfg: EncryptConfig{Algorithm: AES128, Mode: CBC, Key: "1234567890123456", IV: "abcdefghijklmnop"}},
	{name: "AES128_CBC_RandomIV", cfg: EncryptConfig{Algorithm: AES128, Mode: CBC, Key: "1234567890123456"}},
	{name: "AES128_ECB", cfg: EncryptConfig{Algorithm: AES128, Mode: ECB, Key: "1234567890123456"}},
	{name: "AES128_GCM", cfg: EncryptConfig{Algorithm: AES128, Mode: GCM, Key: "1234567890123456", IV: "abcdefghijkl"}},
	{name: "AES128_GCM_RandomNonce", cfg: EncryptConfig{Algorithm: AES128, Mode: GCM, Key: "1234567890123456"}},
	{name: "AES192_CBC", cfg: EncryptConfig{Algorithm: AES192, Mode: CBC, Key: "123456789012345678901234", IV: "abcdefghijklmnop"}},
	{name: "AES192_ECB", cfg: EncryptConfig{Algorithm: AES192, Mode: ECB, Key: "123456789012345678901234"}},
	{name: "AES192_GCM", cfg: EncryptConfig{Algorithm: AES192, Mode: GCM, Key: "123456789012345678901234"}},
	{name: "AES256_CBC", cfg: EncryptConfig{Algorithm: AES256, Mode: CBC, Key: "12345678901234567890123456789012", IV: "abcdefghijklmnop"}},
	{name: "AES256_ECB", cfg: EncryptConfig{Algorithm: AES256, Mode: ECB, Key: "12345678901234567890123456789012"}},
	{name: "AES256_GCM", cfg: EncryptConfig{Algorithm: AES256, Mode: GCM, Key: "12345678901234567890123456789012"}},
	{name: "Default_CBC", cfg: EncryptConfig{Key: "1234567890123456"}}, // 默认 AES128+CBC
}

func TestEncrypt_AllCombinations(t *testing.T) {
	plaintext := []byte(`{"body":"测试内容","sound":"birdsong"}`)
	for _, tc := range encryptCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			ciphertext, iv, err := encrypt(plaintext, &cfg)
			if err != nil {
				t.Fatalf("加密失败: %v", err)
			}
			if len(ciphertext) == 0 {
				t.Fatal("密文不应为空")
			}
			encoded := base64.StdEncoding.EncodeToString(ciphertext)
			if _, err = base64.StdEncoding.DecodeString(encoded); err != nil {
				t.Fatalf("密文 Base64 解码失败: %v", err)
			}
			// ECB 模式无 IV，其他模式有 IV/Nonce
			if tc.cfg.Mode != ECB && iv == "" {
				t.Fatal("CBC/GCM 模式应返回非空 IV/Nonce")
			}
			if tc.cfg.Mode == ECB && iv != "" {
				t.Fatal("ECB 模式不应返回 IV")
			}
		})
	}
}

// --- 集成测试（需要有效的 DeviceKey） ---

func TestNotify_Basic(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "基础推送",
		Body:      "TestNotify_Basic",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_OnlyBody(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Body:      "只有内容，没有标题",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Sound(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "铃声测试",
		Body:      "TestNotify_Sound",
		Sound:     "paymentsuccess",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Call(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "持续响铃",
		Body:      "TestNotify_Call",
		Call:      true,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Icon(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "自定义图标",
		Body:      "TestNotify_Icon",
		Icon:      "https://day.app/assets/images/avatar.jpg",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Group(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "消息分组",
		Body:      "TestNotify_Group",
		Group:     "测试分组",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_NotArchive(t *testing.T) {
	f := false
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "不存档",
		Body:      "TestNotify_NotArchive",
		IsArchive: &f,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Level_Active(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "普通通知",
		Body:      "TestNotify_Level_Active",
		Level:     Active,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Level_TimeSensitive(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "时效性通知",
		Body:      "TestNotify_Level_TimeSensitive",
		Level:     TimeSensitive,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Level_Passive(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "静默通知",
		Body:      "TestNotify_Level_Passive",
		Level:     Passive,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Level_Critical(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "重要警告",
		Body:      "TestNotify_Level_Critical",
		Level:     Critical,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Critical_CustomVolume(t *testing.T) {
	vol := 8
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "重要警告（自定义音量）",
		Body:      "TestNotify_Critical_CustomVolume",
		Level:     Critical,
		Volume:    &vol,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_URL(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "URL 跳转",
		Body:      "TestNotify_URL（点击跳转 GitHub）",
		URL:       "https://github.com/lkzc19/bark.sdk.go",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Copy(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "复制内容",
		Body:      "TestNotify_Copy",
		Copy:      "https://pkg.go.dev/github.com/lkzc19/bark.sdk.go",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_AutoCopy(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "自动复制",
		Body:      "TestNotify_AutoCopy（Body 内容会自动复制）",
		AutoCopy:  true,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Badge(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "角标",
		Body:      "TestNotify_Badge",
		Badge:     42,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_SpecialChars(t *testing.T) {
	key := before(t)
	err := Notify(Req{
		DeviceKey: key,
		Title:     "特殊字符 & 中文",
		Body:      "CPU 使用率超过 90%！服务器 IP: 192.168.1.1",
		Group:     "服务器监控",
		URL:       "https://monitor.example.com?host=web&port=443",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Encrypt_FixedIV(t *testing.T) {
	deviceKey := before(t)
	cryptoKey := os.Getenv("CryptoKey")
	cryptoIV := os.Getenv("CryptoIV")
	err := Notify(Req{
		DeviceKey: deviceKey,
		Title:     "加密推送（AES128-CBC，固定 IV）",
		Body:      "TestNotify_Encrypt_FixedIV",
		AutoCopy:  true,
		Encrypt:   &EncryptConfig{Algorithm: AES128, Mode: CBC, Key: cryptoKey, IV: cryptoIV},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Encrypt_RandomIV(t *testing.T) {
	deviceKey := before(t)
	cryptoKey := os.Getenv("CryptoKey")
	if cryptoKey == "" || len(cryptoKey) != 16 {
		t.Skip("环境变量 CryptoKey 未配置或不足 16 字节，跳过加密集成测试")
	}
	err := Notify(Req{
		DeviceKey: deviceKey,
		Title:     "加密推送（AES128-CBC，随机 IV）",
		Body:      "TestNotify_Encrypt_RandomIV",
		Encrypt:   &EncryptConfig{Algorithm: AES128, Mode: CBC, Key: cryptoKey},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestNotify_Encrypt_AES256_GCM(t *testing.T) {
	deviceKey := before(t)
	cryptoKey := os.Getenv("CryptoKey256")
	if cryptoKey == "" || len(cryptoKey) != 32 {
		t.Skip("环境变量 CryptoKey256 未配置或不足 32 字节，跳过 AES256-GCM 集成测试")
	}
	err := Notify(Req{
		DeviceKey: deviceKey,
		Title:     "加密推送（AES256-GCM）",
		Body:      "TestNotify_Encrypt_AES256_GCM",
		Encrypt:   &EncryptConfig{Algorithm: AES256, Mode: GCM, Key: cryptoKey},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestClient_NewWithURL(t *testing.T) {
	// 验证自定义 URL 客户端构造正常
	client := NewWithURL("https://custom.bark.server")
	if client.baseURL != "https://custom.bark.server" {
		t.Fatalf("期望 baseURL=https://custom.bark.server，实际 %s", client.baseURL)
	}
}
