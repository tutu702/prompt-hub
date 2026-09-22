// Package password 提供基于 bcrypt 的密码哈希与校验能力。
//
// 骨架阶段使用 bcrypt.DefaultCost（10），约 50ms/次；如需更高强度，
// 后续可在环境变量中暴露 cost 配置，本期不做。
package password

import "golang.org/x/crypto/bcrypt"

// Hash 返回明文密码的 bcrypt 哈希。
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify 校验明文密码与哈希是否匹配。
func Verify(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}