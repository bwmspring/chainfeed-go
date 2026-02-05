package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/sha3"
)

var addressRegex = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

// IsHexAddress 验证是否为有效的以太坊地址
func IsHexAddress(s string) bool {
	return addressRegex.MatchString(s)
}

// HexToAddress 标准化地址格式
func HexToAddress(s string) string {
	if !strings.HasPrefix(s, "0x") {
		s = "0x" + s
	}
	return strings.ToLower(s)
}

// Keccak256 计算 Keccak256 哈希
func Keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// RecoverAddress 从签名恢复地址
func RecoverAddress(message []byte, signature []byte) (string, error) {
	if len(signature) != 65 {
		return "", errors.New("invalid signature length")
	}

	// 调整 v 值
	if signature[64] == 27 || signature[64] == 28 {
		signature[64] -= 27
	}

	// 使用 secp256k1 恢复公钥
	pubKey, err := recoverPubkey(message, signature)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key: %w", err)
	}

	// 从公钥生成地址
	address := PubkeyToAddress(pubKey)
	return address, nil
}

// PubkeyToAddress 从公钥生成地址
func PubkeyToAddress(pubKey *ecdsa.PublicKey) string {
	// 序列化公钥（去掉第一个字节的前缀）
	pubBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)

	// Keccak256 哈希
	hash := Keccak256(pubBytes)

	// 取后 20 字节作为地址
	address := "0x" + hex.EncodeToString(hash[12:])
	return strings.ToLower(address)
}

// recoverPubkey 从签名恢复公钥（简化实现）
func recoverPubkey(hash []byte, sig []byte) (*ecdsa.PublicKey, error) {
	// 这里需要使用 secp256k1 库来恢复公钥
	// 由于依赖问题，暂时返回错误
	return nil, errors.New("signature recovery not implemented - requires go-ethereum dependency")
}
