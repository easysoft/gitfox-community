package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Extractor 字符串提取器
type Extractor struct {
	config *ExtractorConfig
}

// ExtractorConfig 提取器配置
type ExtractorConfig struct {
	// FunctionNames 要提取的函数名列表
	FunctionNames []string
	// IgnorePatterns 忽略的字符串模式
	IgnorePatterns []*regexp.Regexp
	// IncludeComments 是否包含注释中的字符串
	IncludeComments bool
	// MinLength 最小字符串长度
	MinLength int
	// MaxLength 最大字符串长度
	MaxLength int
}

// ExtractedString 提取的字符串信息
type ExtractedString struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Function string `json:"function,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

// NewExtractor 创建字符串提取器
func NewExtractor(config *ExtractorConfig) *Extractor {
	if config == nil {
		config = &ExtractorConfig{}
	}

	if len(config.FunctionNames) == 0 {
		config.FunctionNames = []string{"T", "Tf", "TWithLocale", "TWithContext", "TfWithContext"}
	}

	if config.MinLength == 0 {
		config.MinLength = 1
	}

	if config.MaxLength == 0 {
		config.MaxLength = 1000
	}

	return &Extractor{
		config: config,
	}
}

// ExtractFromFile 从文件提取字符串
func (e *Extractor) ExtractFromFile(filePath string) ([]*ExtractedString, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse file: %s", filePath)
	}

	var extracted []*ExtractedString

	// 遍历 AST 节点
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			// 检查函数调用
			if ident, ok := x.Fun.(*ast.Ident); ok {
				if e.isFunctionToExtract(ident.Name) {
					strings := e.extractFromCallExpr(fset, filePath, x)
					extracted = append(extracted, strings...)
				}
			}
		case *ast.BasicLit:
			// 检查字符串字面量
			if x.Kind == token.STRING {
				str := e.extractFromBasicLit(fset, filePath, x)
				if str != nil {
					extracted = append(extracted, str)
				}
			}
		}
		return true
	})

	return extracted, nil
}

// ExtractFromDir 从目录提取字符串
func (e *Extractor) ExtractFromDir(dirPath string) (map[string][]*ExtractedString, error) {
	// 这里可以实现目录遍历和文件提取
	// 为了简化，暂时返回空结果
	return make(map[string][]*ExtractedString), nil
}

// isFunctionToExtract 检查是否为要提取的函数
func (e *Extractor) isFunctionToExtract(name string) bool {
	for _, funcName := range e.config.FunctionNames {
		if name == funcName {
			return true
		}
	}
	return false
}

// extractFromCallExpr 从函数调用表达式提取字符串
func (e *Extractor) extractFromCallExpr(fset *token.FileSet, filePath string, call *ast.CallExpr) []*ExtractedString {
	var extracted []*ExtractedString

	// 提取第一个参数（通常是 key）
	if len(call.Args) > 0 {
		if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			str := e.extractFromBasicLit(fset, filePath, lit)
			if str != nil {
				// 设置函数名
				if ident, ok := call.Fun.(*ast.Ident); ok {
					str.Function = ident.Name
				}
				extracted = append(extracted, str)
			}
		}
	}

	return extracted
}

// extractFromBasicLit 从基本字面量提取字符串
func (e *Extractor) extractFromBasicLit(fset *token.FileSet, filePath string, lit *ast.BasicLit) *ExtractedString {
	// 解析字符串值
	value := strings.Trim(lit.Value, `"`)

	// 检查长度限制
	if len(value) < e.config.MinLength || len(value) > e.config.MaxLength {
		return nil
	}

	// 检查忽略模式
	for _, pattern := range e.config.IgnorePatterns {
		if pattern.MatchString(value) {
			return nil
		}
	}

	// 获取位置信息
	pos := fset.Position(lit.Pos())

	return &ExtractedString{
		Key:    generateKey(value),
		Value:  value,
		File:   filePath,
		Line:   pos.Line,
		Column: pos.Column,
	}
}

// generateKey 生成翻译键
func generateKey(value string) string {
	// 简单的键生成策略：使用原始值作为键
	// 可以根据需要实现更复杂的键生成逻辑
	key := strings.ToLower(value)
	key = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(key, "_")
	key = strings.Trim(key, "_")

	if len(key) > 50 {
		key = key[:50]
	}

	return key
}
