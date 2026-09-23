// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/logger"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// - 设置 SSH 服务器的主机密钥
// - 如果没有提供密钥，会创建默认密钥
// - 将密钥配置到内部 SSH 服务器
func (s *Server) setupHostKeys() error {
	keys := make([]string, 0, len(s.HostKeys))
	for _, key := range s.HostKeys {
		_, err := os.Stat(key)
		if err != nil {
			return fmt.Errorf("failed to read provided host key %q: %w", key, err)
		}
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		//log.Debug().Msg("no host key provided - setup default key if it doesn't exist yet")
		keyPath := filepath.Join(s.ServerHostKeysDir, defaultServerKeyPath)
		err := createKeyIfNotExists(keyPath)
		if err != nil {
			return fmt.Errorf("failed to setup default key %q: %w", keyPath, err)
		}
		keys = append(keys, keyPath)
	}

	// set keys to internal ssh server
	for _, key := range keys {
		err := s.internal.SetOption(ssh.HostKeyFile(key))
		if err != nil {
			logger.Errorw("failed to set host key to ssh server", "error", err)
		}
	}

	return nil
}

// - 检查密钥文件是否存在，不存在则创建
// - 创建必要的目录结构
func createKeyIfNotExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		// if the path already exists there's nothing we have to do
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to check for for existence of key: %w", err)
	}

	logger.Debugf("generate new key at %q", path)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create dir %q for key: %w", dir, err)
	}

	err = GenerateKeyPair(path)
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	return nil
}

// GenerateKeyPair make a pair of public and private keys for SSH access.
// - 生成 RSA 密钥对（4096位）
// - 创建私钥文件（.pem 格式）
// - 创建公钥文件（.pub 格式）
func GenerateKeyPair(keyPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	f, err := os.OpenFile(keyPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := pem.Encode(f, privateKeyPEM); err != nil {
		return err
	}

	// generate public key
	pub, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	public := gossh.MarshalAuthorizedKey(pub)
	p, err := os.OpenFile(keyPath+".pub", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer p.Close()

	_, err = p.Write(public)
	if err != nil {
		return fmt.Errorf("failed to write to public key: %w", err)
	}
	return nil
}
