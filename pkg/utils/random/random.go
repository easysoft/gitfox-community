package random

import (
	"encoding/base32"
	"errors"
	"math"
	"math/rand"
	"os"
	"time"
	"unsafe"
)

const (
	MaximumCapacity = math.MaxInt32>>1 + 1
	Numeral         = "0123456789"
	LowwerLetters   = "abcdefghijklmnopqrstuvwxyz"
	UpperLetters    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Letters         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	SymbolChars     = "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	AllChars        = Numeral + LowwerLetters + UpperLetters + SymbolChars
)

var rn = rand.NewSource(time.Now().UnixNano())

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandInt generate random int between [min, max).
// 生成[min， max]之间的随机整数。
func RandInt(min, max int) int {
	if min == max {
		return min
	}

	if max < min {
		min, max = max, min
	}

	if min == 0 && max == math.MaxInt {
		return rand.Int()
	}

	return rand.Intn(max-min) + min
}

// RandString generate random alphabeta string of specified length.
// 生成指定长度的随机字母数字字符串。
func RandString(length int) string {
	return random(Letters, length)
}

// RandNumeral generate a random numeral string of specified length.
// 生成指定长度的随机数字字符串。
func RandNumeral(length int) string {
	return random(Numeral, length)
}

// RandNumeralOrLetter generate a random numeral or alpha string of specified length.
// 生成指定长度的随机数字或字母字符串。
func RandNumeralOrLetter(length int) string {
	return random(Numeral+Letters, length)
}

// RandSymbolChar generate a random symbol char of specified length.
// symbol chars: !@#$%^&*()_+-=[]{}|;':\",./<>?.
// 生成指定长度的随机符号字符字符串。
func RandSymbolChar(length int) string {
	return random(SymbolChars, length)
}

type UIDLength int

// UIDLength constants define the length of the generated UID string.
// Note: These values represent the number of characters in the output string,
// NOT the CPU architecture (e.g., 32-bit or 64-bit processor).
// It works consistently on any architecture.
const (
	// UIDLength16 generates a 16-char string.
	// Example: "ORSXG5A6NZQXG2L"
	UIDLength16 UIDLength = 16

	// UIDLength32 generates a 32-char string.
	// Example: "ORSXG5A6NZQXG2LORSXG5A6NZQXG2L12"
	UIDLength32 UIDLength = 32

	// UIDLength64 generates a 64-char string.
	// Example: "ORSXG5A6NZQXG2LORSXG5A6NZQXG2L12ORSXG5A6NZQXG2LORSXG5A6NZQXG2L12"
	UIDLength64 UIDLength = 64
)

// UID returns unique random string with the specified length enum.
func UID(l UIDLength) (string, error) {
	length := int(l)
	switch l {
	case UIDLength16, UIDLength32, UIDLength64:
		// valid
	default:
		return "", errors.New("invalid UID length")
	}

	// Calculate number of bytes required to generate at least 'length' base32 characters.
	// Each base32 char is 5 bits. Total bits = length * 5.
	// Total bytes = ceil(total bits / 8).
	uidSizeBytes := (length*5 + 7) / 8

	buf := make([]byte, uidSizeBytes)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	uid := base32.StdEncoding.EncodeToString(buf)

	// base32 encoding might produce more characters (padding or byte alignment).
	// Truncate to the requested length.
	if len(uid) > length {
		uid = uid[:length]
	}

	return uid, nil
}

// random generate a random string based on given string range.
// 基于给定的字符串范围生成随机字符串。
func random(s string, length int) string {
	// 确保随机数生成器的种子是动态的
	pid := os.Getpid()
	timestamp := time.Now().UnixNano()
	rand.Seed(int64(pid) + timestamp)

	// 仿照strings.Builder
	// 创建一个长度为 length 的字节切片
	bytes := make([]byte, length)
	strLength := len(s)
	if strLength <= 0 {
		return ""
	} else if strLength == 1 {
		for i := 0; i < length; i++ {
			bytes[i] = s[0]
		}
		return *(*string)(unsafe.Pointer(&bytes))
	}
	// s的字符需要使用多少个比特位数才能表示完
	// letterIdBits := int(math.Ceil(math.Log2(strLength))),下面比上面的代码快
	letterIdBits := int(math.Log2(float64(nearestPowerOfTwo(strLength))))
	// 最大的字母id掩码
	var letterIdMask int64 = 1<<letterIdBits - 1
	// 可用次数的最大值
	letterIdMax := 63 / letterIdBits
	// 循环生成随机字符串
	for i, cache, remain := length-1, rn.Int63(), letterIdMax; i >= 0; {
		// 检查随机数生成器是否用尽所有随机数
		if remain == 0 {
			cache, remain = rn.Int63(), letterIdMax
		}
		// 从可用字符的字符串中随机选择一个字符
		if idx := int(cache & letterIdMask); idx < strLength {
			bytes[i] = s[idx]
			i--
		}
		// 右移比特位数，为下次选择字符做准备
		cache >>= letterIdBits
		remain--
	}
	// 仿照strings.Builder用unsafe包返回一个字符串，避免拷贝
	// 将字节切片转换为字符串并返回
	return *(*string)(unsafe.Pointer(&bytes))
}

// nearestPowerOfTwo 返回一个大于等于cap的最近的2的整数次幂，参考java8的hashmap的tableSizeFor函数
func nearestPowerOfTwo(cap int) int {
	n := cap - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if n < 0 {
		return 1
	} else if n >= MaximumCapacity {
		return MaximumCapacity
	}
	return n + 1
}
