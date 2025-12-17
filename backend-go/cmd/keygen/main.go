package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// RSA 密钥生成工具
// 用法：go run cmd/keygen/main.go [output_path]
// 默认输出到 ./keys/private.pem

func main() {
	outputPath := "./keys/private.pem"
	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}

	fmt.Println("🔑 Generating RSA key pair (2048 bits)...")

	// 生成 2048 位 RSA 密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("❌ Failed to generate key: %v\n", err)
		os.Exit(1)
	}

	// 将私钥编码为 PEM 格式
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	// 保存到文件（权限 600：仅所有者可读写）
	if err := os.WriteFile(outputPath, keyPEM, 0600); err != nil {
		fmt.Printf("❌ Failed to save key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Private key saved to: %s\n", outputPath)
	fmt.Println("⚠️  IMPORTANT: Keep this file secure and never commit it to Git!")
	fmt.Printf("   File permissions: 0600 (read/write for owner only)\n")
}
