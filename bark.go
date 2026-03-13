package bark

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const defaultBaseURL = "https://api.day.app"

// Level 通知优先级
type Level string

const (
	// Active 普通通知，系统会立即亮屏显示（默认值）
	Active Level = "active"
	// TimeSensitive 时效性通知，可在专注模式下显示
	TimeSensitive Level = "timeSensitive"
	// Passive 仅将通知添加到通知列表，不会亮屏提醒
	Passive Level = "passive"
	// Critical 重要警告，会忽略静音设置和勿扰模式
	Critical Level = "critical"
)

// Algorithm AES 密钥长度（决定 AES 位数）
type Algorithm string

const (
	AES128 Algorithm = "AES128" // 密钥 16 字节
	AES192 Algorithm = "AES192" // 密钥 24 字节
	AES256 Algorithm = "AES256" // 密钥 32 字节
)

// Mode 加密模式
type Mode string

const (
	// CBC 需要 IV（16 字节），PKCS7 填充
	CBC Mode = "CBC"
	// ECB 无需 IV，PKCS7 填充（不推荐用于安全敏感场景）
	ECB Mode = "ECB"
	// GCM 需要 Nonce（12 字节），无需填充，含消息认证标签
	GCM Mode = "GCM"
)

// algorithmKeyLen 返回各算法对应的密钥字节数
var algorithmKeyLen = map[Algorithm]int{
	AES128: 16,
	AES192: 24,
	AES256: 32,
}

// modeIVLen 返回各模式对应的 IV/Nonce 字节数（0 表示不需要）
var modeIVLen = map[Mode]int{
	CBC: 16,
	ECB: 0,
	GCM: 12,
}

// EncryptConfig 端到端加密配置
// 需在 Bark App「设置 → 加密」中配置相同的算法与密钥
type EncryptConfig struct {
	// Algorithm 密钥长度，默认 AES128
	Algorithm Algorithm
	// Mode 加密模式，默认 CBC
	Mode Mode
	// Key 加密密钥（长度由 Algorithm 决定：AES128=16字节，AES192=24字节，AES256=32字节）
	Key string
	// IV CBC 模式的初始向量（16字节）或 GCM 模式的 Nonce（12字节）
	// ECB 模式忽略此字段；其他模式为空时自动随机生成
	IV string
}

// Req 推送请求参数
type Req struct {
	// DeviceKey 设备唯一标识，打开 Bark App 首页获取（必填）
	DeviceKey string
	// Title 推送标题（与 Body 至少填一个）
	Title string
	// Body 推送内容（与 Title 至少填一个）
	Body string
	// Sound 自定义铃声名称，默认 default
	Sound string
	// Call 持续响铃，持续重复约 30 秒
	Call bool
	// IsArchive 是否存档消息，nil 时使用默认值（存档），true = 存档，false = 不存档
	IsArchive *bool
	// Icon 自定义推送图标 URL
	Icon string
	// Group 消息分组名称
	Group string
	// Level 通知优先级
	Level Level
	// Volume 重要警告音量，仅 Level=Critical 时有效，范围 0~10，默认 5
	Volume *int
	// URL 点击推送时跳转的链接
	URL string
	// Copy 点击复制按钮时复制的内容
	Copy string
	// AutoCopy 是否自动复制 Body 内容
	AutoCopy bool
	// Badge 角标数字
	Badge int
	// Encrypt 加密配置，设置后推送内容将加密传输
	Encrypt *EncryptConfig
}

// postBody 普通推送 POST 请求体（与 Bark API 字段对应）
type postBody struct {
	Title     string `json:"title,omitempty"`
	Body      string `json:"body,omitempty"`
	Sound     string `json:"sound,omitempty"`
	Call      int    `json:"call,omitempty"`
	IsArchive *int   `json:"isArchive,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Group     string `json:"group,omitempty"`
	Level     Level  `json:"level,omitempty"`
	Volume    *int   `json:"volume,omitempty"`
	URL       string `json:"url,omitempty"`
	Copy      string `json:"copy,omitempty"`
	AutoCopy  int    `json:"autoCopy,omitempty"`
	Badge     int    `json:"badge,omitempty"`
}

// encryptedBody 加密推送 POST 请求体
type encryptedBody struct {
	Ciphertext string `json:"ciphertext"`
	IV         string `json:"iv,omitempty"` // ECB 模式时为空
}

type _resp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// Client Bark 推送客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New 创建使用官方服务端的默认客户端
func New() *Client {
	return &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
	}
}

// NewWithURL 创建使用自建服务端的客户端（私有化部署场景）
func NewWithURL(serverURL string) *Client {
	return &Client{
		baseURL:    serverURL,
		httpClient: &http.Client{},
	}
}

// Notify 发送推送通知
func (c *Client) Notify(req Req) error {
	if err := validateReq(req); err != nil {
		return err
	}
	if req.Encrypt != nil {
		return c.sendEncrypted(req)
	}
	return c.send(req)
}

func validateReq(req Req) error {
	if req.DeviceKey == "" {
		return errors.New("参数[DeviceKey]不可为空")
	}
	if req.Title == "" && req.Body == "" {
		return errors.New("参数[Title Body]至少需要一个")
	}
	if req.Level != "" &&
		req.Level != Active &&
		req.Level != TimeSensitive &&
		req.Level != Passive &&
		req.Level != Critical {
		return fmt.Errorf("参数[Level]值非法: %s", req.Level)
	}
	if req.Encrypt != nil {
		return validateEncryptConfig(req.Encrypt)
	}
	return nil
}

func validateEncryptConfig(cfg *EncryptConfig) error {
	algo := cfg.Algorithm
	if algo == "" {
		algo = AES128
	}
	keyLen, ok := algorithmKeyLen[algo]
	if !ok {
		return fmt.Errorf("参数[Encrypt.Algorithm]值非法: %s，可选值: AES128 AES192 AES256", algo)
	}
	if len(cfg.Key) != keyLen {
		return fmt.Errorf("参数[Encrypt.Key]在 %s 模式下必须为 %d 字节，当前 %d 字节", algo, keyLen, len(cfg.Key))
	}

	mode := cfg.Mode
	if mode == "" {
		mode = CBC
	}
	ivLen, ok := modeIVLen[mode]
	if !ok {
		return fmt.Errorf("参数[Encrypt.Mode]值非法: %s，可选值: CBC ECB GCM", mode)
	}
	if ivLen > 0 && cfg.IV != "" && len(cfg.IV) != ivLen {
		return fmt.Errorf("参数[Encrypt.IV]在 %s 模式下必须为 %d 字节，当前 %d 字节", mode, ivLen, len(cfg.IV))
	}
	return nil
}

// send 发送普通（明文）推送
func (c *Client) send(req Req) error {
	endpoint := fmt.Sprintf("%s/%s", c.baseURL, url.PathEscape(req.DeviceKey))
	return c.doPost(endpoint, buildPostBody(req))
}

// sendEncrypted 发送加密推送
// 将完整推送参数 JSON 加密后，仅发送 ciphertext 和 iv
func (c *Client) sendEncrypted(req Req) error {
	plaintext, err := json.Marshal(buildPostBody(req))
	if err != nil {
		return fmt.Errorf("序列化推送内容失败: %w", err)
	}

	ciphertext, iv, err := encrypt(plaintext, req.Encrypt)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/%s", c.baseURL, url.PathEscape(req.DeviceKey))
	return c.doPost(endpoint, encryptedBody{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		IV:         iv,
	})
}

// encrypt 根据配置加密数据，返回密文字节和 IV 字符串
func encrypt(plaintext []byte, cfg *EncryptConfig) (ciphertext []byte, iv string, err error) {
	algo := cfg.Algorithm
	if algo == "" {
		algo = AES128
	}
	mode := cfg.Mode
	if mode == "" {
		mode = CBC
	}

	block, err := aes.NewCipher([]byte(cfg.Key))
	if err != nil {
		return nil, "", fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	switch mode {
	case CBC:
		return encryptCBC(block, plaintext, cfg.IV)
	case ECB:
		return encryptECB(block, plaintext)
	case GCM:
		return encryptGCM(block, plaintext, cfg.IV)
	default:
		return nil, "", fmt.Errorf("不支持的加密模式: %s", mode)
	}
}

func encryptCBC(block cipher.Block, plaintext []byte, ivStr string) ([]byte, string, error) {
	iv := make([]byte, aes.BlockSize)
	if ivStr != "" {
		iv = []byte(ivStr)
	} else {
		if _, err := rand.Read(iv); err != nil {
			return nil, "", fmt.Errorf("生成随机 IV 失败: %w", err)
		}
	}
	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)
	return ciphertext, string(iv), nil
}

func encryptECB(block cipher.Block, plaintext []byte) ([]byte, string, error) {
	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	bs := block.BlockSize()
	for i := 0; i < len(padded); i += bs {
		block.Encrypt(ciphertext[i:i+bs], padded[i:i+bs])
	}
	return ciphertext, "", nil
}

func encryptGCM(block cipher.Block, plaintext []byte, nonceStr string) ([]byte, string, error) {
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", fmt.Errorf("创建 GCM cipher 失败: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize()) // 12 字节
	if nonceStr != "" {
		nonce = []byte(nonceStr)
	} else {
		if _, err = rand.Read(nonce); err != nil {
			return nil, "", fmt.Errorf("生成随机 Nonce 失败: %w", err)
		}
	}
	// Seal 追加认证标签到密文尾部
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, string(nonce), nil
}

// doPost 发送 POST JSON 请求并处理响应
func (c *Client) doPost(endpoint string, body any) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var barkResp _resp
		if jsonErr := json.Unmarshal(respBody, &barkResp); jsonErr != nil {
			return fmt.Errorf("[bark]请求失败(状态码 %d): %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("[bark]请求失败: %s", barkResp.Message)
	}
	return nil
}

func buildPostBody(req Req) postBody {
	b := postBody{
		Title: req.Title,
		Body:  req.Body,
		Sound: req.Sound,
		Icon:  req.Icon,
		Group: req.Group,
		Level: req.Level,
		URL:   req.URL,
		Copy:  req.Copy,
		Badge: req.Badge,
	}

	if req.Call {
		b.Call = 1
	}

	if req.AutoCopy {
		b.AutoCopy = 1
	}

	if req.IsArchive != nil {
		v := 0
		if *req.IsArchive {
			v = 1
		}
		b.IsArchive = &v
	}

	if req.Level == Critical {
		vol := 5
		if req.Volume != nil {
			vol = *req.Volume
		}
		if vol > 10 {
			vol = 10
		}
		if vol < 0 {
			vol = 0
		}
		b.Volume = &vol
	}

	return b
}

// pkcs7Pad 对数据进行 PKCS7 填充，使其对齐到 blockSize
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

// 包级别便捷函数，使用默认客户端（官方服务端）
var defaultClient = New()

// Notify 使用默认客户端发送推送通知
func Notify(req Req) error {
	return defaultClient.Notify(req)
}
