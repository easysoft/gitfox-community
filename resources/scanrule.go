// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package resources

import (
	"encoding/json"
	"fmt"
	"path"
)

// RulesFile 规则文件结构
type RulesFile struct {
	Version string     `json:"version"`
	Source  string     `json:"source"`
	Lang    string     `json:"lang"`
	Plugin  string     `json:"plugin"`
	Rules   []RuleJSON `json:"rules"`
}

// RuleJSON JSON 文件中的规则结构
type RuleJSON struct {
	RuleKey       string `json:"ruleKey"`
	Plugin        string `json:"plugin"`
	Lang          string `json:"lang"`
	Name          string `json:"name"`
	Priority      string `json:"priority"`
	Type          string `json:"type"`
	Tag           string `json:"tag,omitempty"`
	Content       string `json:"content,omitempty"`
	ContentEn     string `json:"content_en,omitempty"`
	Description   string `json:"description,omitempty"`
	DescriptionEn string `json:"description_en,omitempty"`
}

// LoadAllRulesFiles 从嵌入的文件系统中加载所有规则文件
func LoadAllRulesFiles() ([]*RulesFile, error) {
	// 获取所有 *.rules.json 文件
	entries, err := EmbeddedScanRules.ReadDir("scanrule")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded scanrule directory: %w", err)
	}

	var result []*RulesFile

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只处理 *.rules.json 文件
		filename := entry.Name()
		if len(filename) < 12 || filename[len(filename)-11:] != ".rules.json" {
			continue
		}

		// 读取文件内容
		filePath := path.Join("scanrule", filename)
		data, err := EmbeddedScanRules.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		// 解析 JSON
		var rulesFile RulesFile
		if err = json.Unmarshal(data, &rulesFile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal file %s: %w", filename, err)
		}

		result = append(result, &rulesFile)
	}

	return result, nil
}
