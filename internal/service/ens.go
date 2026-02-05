package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	ens "github.com/wealdtech/go-ens/v3"
)

type ENSService struct {
	client *ethclient.Client
}

func NewENSService(rpcURL string) (*ENSService, error) {
	if rpcURL == "" {
		return nil, fmt.Errorf("rpc url is required")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum: %w", err)
	}

	return &ENSService{client: client}, nil
}

// Resolve ENS 名称 -> 地址
func (s *ENSService) Resolve(ctx context.Context, ensName string) (string, error) {
	if !strings.HasSuffix(ensName, ".eth") {
		return "", fmt.Errorf("invalid ENS name: must end with .eth")
	}

	resolver, err := ens.NewResolver(s.client, ensName)
	if err != nil {
		return "", fmt.Errorf("failed to get ENS resolver: %w", err)
	}

	address, err := resolver.Address()
	if err != nil {
		return "", fmt.Errorf("failed to resolve ENS name: %w", err)
	}

	return address.Hex(), nil
}

// ReverseResolve 地址 -> ENS 名称
func (s *ENSService) ReverseResolve(ctx context.Context, address string) (string, error) {
	addr := common.HexToAddress(address)
	name, err := ens.ReverseResolve(s.client, addr)
	if err != nil {
		// 反向解析失败不是错误，可能该地址没有设置 ENS
		return "", nil
	}
	return name, nil
}

func (s *ENSService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

