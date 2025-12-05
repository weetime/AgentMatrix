package kit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
)

const (
	// SM2PublicKeyParam SM2公钥参数代码
	SM2PublicKeyParam = "server.public_key"
	// SM2PrivateKeyParam SM2私钥参数代码
	SM2PrivateKeyParam = "server.private_key"
	// SM2CaptchaLength 验证码长度（嵌入在密码中的前5位）
	SM2CaptchaLength = 5
)

// SM2Encrypt SM2加密
// publicKeyHex: 十六进制格式的公钥（未压缩格式，04开头，65字节）
// plaintext: 明文数据
// 返回: 十六进制格式的密文
func SM2Encrypt(publicKeyHex string, plaintext string) (string, error) {
	// 解析公钥（十六进制格式，未压缩格式，04开头）
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	// 如果公钥以04开头，去掉04前缀（未压缩格式，65字节 -> 64字节）
	if len(publicKeyBytes) == 65 && publicKeyBytes[0] == 0x04 {
		publicKeyBytes = publicKeyBytes[1:]
	}

	// 公钥应该是64字节（x和y坐标各32字节）
	if len(publicKeyBytes) != 64 {
		return "", fmt.Errorf("invalid public key length: expected 64 bytes, got %d", len(publicKeyBytes))
	}

	// 创建压缩格式的公钥（02或03开头，33字节）
	compressedKey := make([]byte, 33)
	// 根据y坐标的奇偶性选择02或03
	if publicKeyBytes[63]&1 == 1 {
		compressedKey[0] = 0x03 // y坐标是奇数
	} else {
		compressedKey[0] = 0x02 // y坐标是偶数
	}
	copy(compressedKey[1:], publicKeyBytes[:32])

	// 解压缩得到公钥对象
	pubKey := sm2.Decompress(compressedKey)
	if pubKey == nil {
		return "", fmt.Errorf("failed to decompress public key")
	}

	// 加密（使用C1C3C2模式）
	ciphertext, err := sm2.Encrypt(pubKey, []byte(plaintext), rand.Reader, sm2.C1C3C2)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	// 返回十六进制格式的密文（去掉04前缀，与Java版本保持一致）
	ciphertextHex := hex.EncodeToString(ciphertext)
	if len(ciphertextHex) > 2 && ciphertextHex[:2] == "04" {
		ciphertextHex = ciphertextHex[2:]
	}

	return ciphertextHex, nil
}

// SM2Decrypt SM2解密
// privateKeyHex: 十六进制格式的私钥（BigInteger的十六进制表示）
// ciphertextHex: 十六进制格式的密文
// 返回: 明文数据
func SM2Decrypt(privateKeyHex string, ciphertextHex string) (string, error) {
	// 检查输入是否为空
	if ciphertextHex == "" {
		return "", fmt.Errorf("ciphertext is empty")
	}

	// 检查是否为有效的十六进制字符串
	if !isValidHexString(ciphertextHex) {
		return "", fmt.Errorf("invalid ciphertext format: password must be SM2 encrypted (hexadecimal string). Please encrypt the password using SM2 public key first")
	}

	// 如果密文不以04开头，则补上（与Java版本保持一致）
	if len(ciphertextHex) > 0 && len(ciphertextHex) >= 2 && ciphertextHex[:2] != "04" {
		ciphertextHex = "04" + ciphertextHex
	}

	// 解析密文
	ciphertextBytes, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// 解析私钥（BigInteger的十六进制表示）
	privateKeyBigInt, ok := new(big.Int).SetString(privateKeyHex, 16)
	if !ok {
		return "", fmt.Errorf("failed to parse private key as big integer")
	}

	// 创建SM2私钥对象
	privKey := new(sm2.PrivateKey)
	privKey.D = privateKeyBigInt
	privKey.PublicKey.Curve = sm2.P256Sm2()

	// 计算公钥点
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(privKey.D.Bytes())

	// 解密（使用C1C3C2模式）
	plaintext, err := sm2.Decrypt(privKey, ciphertextBytes, sm2.C1C3C2)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// CaptchaService 验证码服务接口
type CaptchaService interface {
	Validate(uuid, code string, delete bool) bool
}

// ParamsService 系统参数服务接口
type ParamsService interface {
	GetValue(paramCode string, isCache bool) (string, error)
}

// DecryptAndValidateCaptcha 解密SM2加密内容，提取验证码并验证
// encryptedPassword: SM2加密的密码字符串（前5位是验证码）
// captchaId: 验证码ID
// captchaService: 验证码服务
// paramsService: 系统参数服务
// 返回: 解密后的实际密码
func DecryptAndValidateCaptcha(
	encryptedPassword string,
	captchaId string,
	captchaService CaptchaService,
	paramsService ParamsService,
) (string, error) {
	// 获取SM2私钥
	privateKeyStr, err := paramsService.GetValue(SM2PrivateKeyParam, true)
	if err != nil {
		return "", fmt.Errorf("failed to get SM2 private key: %w", err)
	}
	if privateKeyStr == "" {
		return "", fmt.Errorf("SM2 private key not configured")
	}

	// 使用SM2私钥解密密码
	decryptedContent, err := SM2Decrypt(privateKeyStr, encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("SM2 decrypt error: %w", err)
	}

	// 检查解密后的内容长度
	if len(decryptedContent) == 0 {
		return "", fmt.Errorf("decrypted content is empty")
	}

	// 分离验证码和密码：前5位是验证码，后面是密码
	if len(decryptedContent) <= SM2CaptchaLength {
		return "", fmt.Errorf("decrypted content too short")
	}

	embeddedCaptcha := decryptedContent[:SM2CaptchaLength]
	actualPassword := decryptedContent[SM2CaptchaLength:]

	// 验证嵌入的验证码是否正确
	if !captchaService.Validate(captchaId, embeddedCaptcha, true) {
		return "", fmt.Errorf("captcha validation failed")
	}

	return actualPassword, nil
}

// isValidHexString 检查字符串是否为有效的十六进制字符串
func isValidHexString(s string) bool {
	if len(s) == 0 {
		return false
	}
	// 十六进制字符串只能包含 0-9, a-f, A-F
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
