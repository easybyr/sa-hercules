package password

import "golang.org/x/crypto/bcrypt"

// Hash 使用bcrypt生成密码哈希。
func Hash(raw string) (string, error) {
	value, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	return string(value), err
}

// Verify 验证密码。
func Verify(hash string, raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}
