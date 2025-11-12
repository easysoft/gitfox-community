package constant

// 常量：和 MySQL varchar(500) 对齐
const (
	MaxDescCharCount        = 500                  // 最大字符数（核心限定）
	MaxDescByteCountUTF8MB4 = MaxDescCharCount * 4 // 2000 字节
)

// 编码
const (
	ContentEncodingTypeUTF8   = "utf8"
	ContentEncodingTypeBase64 = "base64"
)
