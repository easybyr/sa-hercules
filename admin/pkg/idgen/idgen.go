package idgen

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const passwordAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%"

// Numeric 生成指定长度的正整数，首位不会为0。
func Numeric(digits int) (int64, error) {
	if digits < 1 || digits > 18 {
		return 0, fmt.Errorf("数字ID长度必须在1到18之间")
	}
	min := int64(1)
	for index := 1; index < digits; index++ {
		min *= 10
	}
	rangeSize := new(big.Int).SetInt64(min * 9)
	offset, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return 0, err
	}
	return min + offset.Int64(), nil
}

// Password 生成适合作为一次性初始密码的随机字符串。
func Password(length int) (string, error) {
	if length < 8 {
		return "", fmt.Errorf("随机密码长度不能少于8位")
	}
	result := make([]byte, length)
	limit := big.NewInt(int64(len(passwordAlphabet)))
	for index := range result {
		position, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		result[index] = passwordAlphabet[position.Int64()]
	}
	return string(result), nil
}

// PersonalRoleID 根据UID生成稳定的个人角色ID。
// 负数角色ID仅供系统内部的用户直接权限分配使用。
func PersonalRoleID(uid int64) int64 {
	return -uid
}
