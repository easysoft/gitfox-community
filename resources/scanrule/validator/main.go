// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// 定义命令行参数
	dir := flag.String("dir", ".", "Directory containing the scan rule files")
	plugin := flag.String("plugin", "", "Specific plugin to validate (optional, validates all if not specified)")
	flag.Parse()

	if *plugin != "" {
		// 验证单个插件
		typesFile := fmt.Sprintf("%s/%s.types.json", *dir, *plugin)
		rulesFile := fmt.Sprintf("%s/%s.rules.json", *dir, *plugin)

		result := ValidatePlugin(*plugin, typesFile, rulesFile)
		PrintValidateResults([]ValidateResult{result})

		if !result.IsValid || result.ErrorMessage != "" {
			os.Exit(1)
		}
	} else {
		// 验证所有插件
		results, err := ValidateAllPlugins(*dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		PrintValidateResults(results)

		// 如果有任何验证失败，退出码为 1
		for _, result := range results {
			if !result.IsValid || result.ErrorMessage != "" {
				os.Exit(1)
			}
		}
	}

	fmt.Println("\n✅ All validations passed!")
}
