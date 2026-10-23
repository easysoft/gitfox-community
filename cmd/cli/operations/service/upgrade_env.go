// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package service

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const pluginsZipURLEnvKey = "GITFOX_CI_PLUGINS_ZIP_URL"

// envFileChange 保存一次待执行的配置文件修改，以及失败时的原始内容。
type envFileChange struct {
	path       string
	original   []byte
	updated    []byte
	mode       os.FileMode
	backupPath string
}

// preparePluginsZipURLDefault 检查当前生效的插件包地址。解析 URL 并校验 scheme 为 http/https 且 host 非空时，只注释 prod.env 中的赋值行，让 envconfig 使用字段默认值。
func preparePluginsZipURLDefault(workDir string) (*envFileChange, error) {
	envPath := filepath.Join(workDir, ProdEnvFileName)
	info, err := os.Lstat(envPath)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s is a symbolic link, refusing to replace it", ProdEnvFileName)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", ProdEnvFileName)
	}

	original, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", ProdEnvFileName, err)
	}
	envMap, err := godotenv.UnmarshalBytes(original)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", ProdEnvFileName, err)
	}

	value, ok := envMap[pluginsZipURLEnvKey]
	if !ok {
		return nil, nil
	}
	u, perr := url.Parse(strings.TrimSpace(value))
	if perr != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, nil
	}

	updated, changed := commentEnvAssignments(original, pluginsZipURLEnvKey)
	if !changed {
		return nil, nil
	}

	return &envFileChange{
		path:     envPath,
		original: original,
		updated:  updated,
		mode:     info.Mode().Perm(),
	}, nil
}

// commentEnvAssignments 按行注释指定环境变量的所有生效赋值。保留其他配置、注释、缩进以及 LF/CRLF 换行格式。
func commentEnvAssignments(content []byte, key string) ([]byte, bool) {
	lines := bytes.SplitAfter(content, []byte("\n"))
	changed := false
	for i, line := range lines {
		indent, ok := envAssignmentIndent(line, key)
		if !ok {
			continue
		}

		commented := make([]byte, 0, len(line)+2)
		commented = append(commented, line[:indent]...)
		commented = append(commented, '#', ' ')
		commented = append(commented, line[indent:]...)
		lines[i] = commented
		changed = true
	}
	if !changed {
		return content, false
	}
	return bytes.Join(lines, nil), true
}

// envAssignmentIndent 返回未注释赋值的缩进长度。
func envAssignmentIndent(line []byte, key string) (int, bool) {
	bodyEnd := len(line)
	if bodyEnd > 0 && line[bodyEnd-1] == '\n' {
		bodyEnd--
	}
	if bodyEnd > 0 && line[bodyEnd-1] == '\r' {
		bodyEnd--
	}

	body := line[:bodyEnd]
	indent := 0
	for indent < len(body) && isEnvWhitespace(body[indent]) {
		indent++
	}
	if indent == len(body) || body[indent] == '#' {
		return 0, false
	}

	rest := body[indent:]
	if bytes.HasPrefix(rest, []byte("export")) && len(rest) > len("export") && isEnvWhitespace(rest[len("export")]) {
		rest = bytes.TrimLeft(rest[len("export"):], " \t")
	}
	if !bytes.HasPrefix(rest, []byte(key)) {
		return 0, false
	}

	rest = rest[len(key):]
	rest = bytes.TrimLeft(rest, " \t")
	if len(rest) == 0 || (rest[0] != '=' && rest[0] != ':') {
		return 0, false
	}
	return indent, true
}

// isEnvWhitespace 判断字符是否为环境变量赋值允许的行内空白。
func isEnvWhitespace(c byte) bool {
	return c == ' ' || c == '\t'
}

// apply 先备份原始 prod.env，再通过临时文件原子替换为更新后的内容。
func (c *envFileChange) apply() error {
	if c == nil {
		return nil
	}
	backupPath, err := backupEnvFile(c.path, c.original, c.mode)
	if err != nil {
		return fmt.Errorf("backup %s: %w", ProdEnvFileName, err)
	}
	c.backupPath = backupPath

	if err := replaceFileAtomically(c.path, c.updated, c.mode); err != nil {
		return fmt.Errorf("write %s after backup %s: %w", ProdEnvFileName, backupPath, err)
	}
	return nil
}

// restore 将 prod.env 原子恢复为本次变更前的原始内容。
func (c *envFileChange) restore() error {
	if c == nil {
		return nil
	}
	if err := replaceFileAtomically(c.path, c.original, c.mode); err != nil {
		return fmt.Errorf("restore %s: %w", ProdEnvFileName, err)
	}
	return nil
}

// backupEnvFile 将修改前的 prod.env 完整写入唯一备份文件。
// 备份成功后才允许修改正式文件；备份失败时正式文件保持原样。
func backupEnvFile(path string, content []byte, mode os.FileMode) (string, error) {
	pattern := fmt.Sprintf("%s.upgrade-bak-%s-*", filepath.Base(path), time.Now().Format("20060102-150405"))
	backup, err := os.CreateTemp(filepath.Dir(path), pattern)
	if err != nil {
		return "", err
	}
	backupPath := backup.Name()
	removeBackup := true
	defer func() {
		if removeBackup {
			_ = os.Remove(backupPath)
		}
	}()

	if err := writeFileContentsAndSync(backup, content, mode); err != nil {
		_ = backup.Close()
		return "", err
	}
	if err := backup.Close(); err != nil {
		return "", err
	}
	removeBackup = false
	return backupPath, nil
}

// replaceFileAtomically 将内容写入同目录临时文件，完成同步后通过 rename 原子替换目标文件。
func replaceFileAtomically(path string, content []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".upgrade-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpName)
		}
	}()

	if err := writeFileContentsAndSync(tmp, content, mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	removeTmp = false
	return nil
}

// writeFileContentsAndSync 设置文件权限、写入完整内容并同步到磁盘。
func writeFileContentsAndSync(file *os.File, content []byte, mode os.FileMode) error {
	if err := file.Chmod(mode); err != nil {
		return err
	}
	if _, err := file.Write(content); err != nil {
		return err
	}
	return file.Sync()
}
