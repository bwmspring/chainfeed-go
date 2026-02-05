package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type Web3Service struct {
	signMessage string
}

func NewWeb3Service(signMessage string) *Web3Service {
	return &Web3Service{
		signMessage: signMessage,
	}
}

// GenerateNonce 生成随机 nonce
func (s *Web3Service) GenerateNonce() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetSignMessage 获取待签名消息
func (s *Web3Service) GetSignMessage(walletAddress, nonce string) string {
	return fmt.Sprintf("%s\n\nWallet: %s\nNonce: %s", s.signMessage, walletAddress, nonce)
}

// VerifySignature 验证 Web3 签名
func (s *Web3Service) VerifySignature(walletAddress, signature, nonce string) error {
	// 标准化地址
	if !common.IsHexAddress(walletAddress) {
		return errors.New("invalid wallet address")
	}
	address := common.HexToAddress(walletAddress)

	// 构建消息
	message := s.GetSignMessage(address.Hex(), nonce)
	messageHash := s.hashMessage(message)

	// 解码签名
	sig, err := hexutil.Decode(signature)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	// 调整 v 值（MetaMask 签名兼容性）
	if sig[64] == 27 || sig[64] == 28 {
		sig[64] -= 27
	}

	// 恢复公钥
	pubKey, err := crypto.SigToPub(messageHash, sig)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// 验证地址
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	if !strings.EqualFold(recoveredAddr.Hex(), address.Hex()) {
		return errors.New("signature verification failed")
	}

	return nil
}

// hashMessage 使用以太坊标准前缀哈希消息
func (s *Web3Service) hashMessage(message string) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	return crypto.Keccak256([]byte(prefix + message))
}
