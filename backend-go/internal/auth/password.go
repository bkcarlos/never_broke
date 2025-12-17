package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Argon2id 参数（OWASP 推荐）
const (
	Argon2Time    = 1         // 迭代次数
	Argon2Memory  = 64 * 1024 // 64 MB
	Argon2Threads = 4         // 并行度
	Argon2KeyLen  = 32        // 输出长度（32 字节）
	SaltLength    = 16        // Salt 长度
)

// HashPassword 使用 argon2id 哈希密码
// 安全说明：
// 1. 使用 argon2id（抗侧信道攻击，OWASP 推荐）
// 2. 每个密码使用唯一的随机 salt
// 3. 返回格式：$argon2id$v=19$m=65536,t=1,p=4$salt$hash
func HashPassword(password string) (string, error) {
	// 生成随机 salt
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// 使用 argon2id 哈希
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLen,
	)

	// 编码为存储格式
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		Argon2Memory, Argon2Time, Argon2Threads, b64Salt, b64Hash,
	), nil
}

// VerifyPassword 验证密码
// 安全说明：
// 1. 使用恒定时间比较（防止时序攻击）
// 2. 失败时不应泄露具体原因（邮箱不存在 vs 密码错误）
func VerifyPassword(password, hashedPassword string) (bool, error) {
	// 解析存储的哈希
	salt, hash, params, err := decodeArgon2Hash(hashedPassword)
	if err != nil {
		return false, err
	}

	// 使用相同参数哈希输入密码
	inputHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.time,
		params.memory,
		params.threads,
		params.keyLen,
	)

	// 恒定时间比较（防止时序攻击）
	if len(hash) != len(inputHash) {
		return false, nil
	}

	var diff byte
	for i := 0; i < len(hash); i++ {
		diff |= hash[i] ^ inputHash[i]
	}

	return diff == 0, nil
}

// argon2Params argon2 参数
type argon2Params struct {
	memory  uint32
	time    uint32
	threads uint8
	keyLen  uint32
}

// decodeArgon2Hash 解析 argon2 哈希
func decodeArgon2Hash(encodedHash string) (salt, hash []byte, params argon2Params, err error) {
	var version int
	_, err = fmt.Sscanf(
		encodedHash,
		"$argon2id$v=%d$m=%d,t=%d,p=%d$",
		&version,
		&params.memory,
		&params.time,
		&params.threads,
	)
	if err != nil {
		return nil, nil, argon2Params{}, errors.New("invalid hash format")
	}

	// 提取 salt 和 hash
	parts := []byte(encodedHash)
	lastDollar := 0
	dollarCount := 0
	for i, b := range parts {
		if b == '$' {
			dollarCount++
			if dollarCount == 5 {
				lastDollar = i
				break
			}
		}
	}

	saltHashPart := string(parts[lastDollar+1:])
	saltHashParts := []byte(saltHashPart)
	dollarPos := 0
	for i, b := range saltHashParts {
		if b == '$' {
			dollarPos = i
			break
		}
	}

	b64Salt := string(saltHashParts[:dollarPos])
	b64Hash := string(saltHashParts[dollarPos+1:])

	salt, err = base64.RawStdEncoding.DecodeString(b64Salt)
	if err != nil {
		return nil, nil, argon2Params{}, errors.New("invalid salt encoding")
	}

	hash, err = base64.RawStdEncoding.DecodeString(b64Hash)
	if err != nil {
		return nil, nil, argon2Params{}, errors.New("invalid hash encoding")
	}

	params.keyLen = uint32(len(hash))

	return salt, hash, params, nil
}
