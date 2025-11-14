// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/controller/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"

	"github.com/easysoft/gitfox/errors"

	"github.com/gliderlabs/ssh"
	"github.com/rs/zerolog/log"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/exp/slices"
)

type Server struct {
	internal *ssh.Server

	Host        string
	Port        int
	DefaultUser string

	TrustedUserCAKeys       []string
	TrustedUserCAKeysParsed []gossh.PublicKey
	Ciphers                 []string
	KeyExchanges            []string
	MACs                    []string
	HostKeys                []string
	KeepAliveInterval       time.Duration
	ServerHostKeysDir       string

	Verifier publickey.IPublicKey
	RepoCtrl *repo.Controller
}

// - 验证服务器配置参数是否正确
// - 设置默认值（如端口、密码学参数等）
// - 检查必要的组件是否已配置（如仓库控制器）
func (s *Server) sanitize() error {
	if s.Port == 0 {
		s.Port = 22
	}

	if len(s.Ciphers) == 0 {
		s.Ciphers = defaultCiphers
	}

	if len(s.KeyExchanges) == 0 {
		s.KeyExchanges = defaultKeyExchanges
	}

	if len(s.MACs) == 0 {
		s.MACs = defaultMACs
	}

	if s.KeepAliveInterval == 0 {
		s.KeepAliveInterval = 5000
	}

	if s.RepoCtrl == nil {
		return errors.InvalidArgument("repository controller is needed to run git service pack commands")
	}
	return nil
}

// - 启动 SSH 服务器的主方法
// - 配置 SSH 服务器参数（地址、处理器、回调函数等）
// - 设置主机密钥并开始监听连接
func (s *Server) ListenAndServe() error {
	err := s.sanitize()
	if err != nil {
		return fmt.Errorf("failed to sanitize server defaults: %w", err)
	}
	s.internal = &ssh.Server{
		Addr:             net.JoinHostPort(s.Host, strconv.Itoa(s.Port)),
		Handler:          s.sessionHandler,
		PublicKeyHandler: s.publicKeyHandler,
		PtyCallback: func(ssh.Context, ssh.Pty) bool {
			return false
		},
		ConnectionFailedCallback: sshConnectionFailed,
		ServerConfigCallback: func(ssh.Context) *gossh.ServerConfig {
			config := &gossh.ServerConfig{}
			config.KeyExchanges = s.KeyExchanges
			config.MACs = s.MACs
			config.Ciphers = s.Ciphers
			return config
		},
	}

	err = s.setupHostKeys()
	if err != nil {
		return fmt.Errorf("failed to setup host keys: %w", err)
	}

	log.Debug().Msgf("starting ssh service....: %v", s.internal.Addr)
	err = s.internal.ListenAndServe()
	if err != nil {
		return fmt.Errorf("ssh service not running: %w", err)
	}
	return nil
}

// 优雅关闭 SSH 服务器
func (s *Server) Shutdown(ctx context.Context) error {
	log.Debug().Msgf("stopping ssh service: %v", s.internal.Addr)
	err := s.internal.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop ssh service: %w", err)
	}
	return nil
}

// - SSH 公钥认证处理器
// - 验证客户端提供的公钥是否有效
// - 检查用户名是否符合要求
// - 支持 SSH 证书认证
// - 将认证成功的用户信息存储到上下文中
func (s *Server) publicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	if slices.Contains(publickey.DisallowedTypes, key.Type()) {
		log.Warn().Msgf("public key type not supported: %s", key.Type())
		return false
	}

	if s.DefaultUser != "" && ctx.User() != s.DefaultUser {
		log.Warn().Msgf("invalid SSH username %s - must use %s for all git operations via ssh",
			ctx.User(), s.DefaultUser)
		log.Warn().Msgf("failed authentication attempt from %s", ctx.RemoteAddr())
		return false
	}

	principal, err := s.Verifier.ValidateKey(ctx, key, constant.PublicKeyUsageAuth)
	if errors.IsNotFound(err) {
		log.Debug().Err(err).Msg("public key is unknown")
		return false
	}
	if err != nil {
		log.Warn().Err(err).Msg("failed to validate public key")
		return false
	}

	// check if we have a certificate
	if cert, ok := key.(*gossh.Certificate); ok {
		if len(s.TrustedUserCAKeys) == 0 {
			log.Warn().Msg("Certificate Rejected: No trusted certificate authorities for this server")
			log.Warn().Msgf("Failed authentication attempt from %s", ctx.RemoteAddr())
			return false
		}

		if cert.CertType != gossh.UserCert {
			log.Warn().Msg("Certificate Rejected: Not a user certificate")
			log.Warn().Msgf("Failed authentication attempt from %s", ctx.RemoteAddr())
			return false
		}

		certChecker := &gossh.CertChecker{}
		if err := certChecker.CheckCert(principal.UID, cert); err != nil {
			return false
		}
	}

	ctx.SetValue(principalKey, principal)
	return true
}

// - 处理 SSH 会话的核心方法
// - 解析 Git 命令（ git-upload-pack 、 git-receive-pack ）
// - 提取仓库路径和协议信息
// - 调用仓库控制器执行 Git 服务包操作pkg/jwt
// - 管理输入输出流
func (s *Server) sessionHandler(session ssh.Session) {
	// eg.git-receive-pack 'space/repo.git'
	command := session.RawCommand()

	principalInfo, ok := session.Context().Value(principalKey).(*principal.PrincipalInfo)
	if !ok {
		_, _ = fmt.Fprintf(session.Stderr(), "principal not found or empty")
		return
	}

	parts := strings.Fields(command)
	if len(parts) < 2 {
		_, _ = fmt.Fprintf(session.Stderr(), "command %q must have an argument\n", command)
		return
	}

	// first part is git service pack command: git-upload-pack, git-receive-pack
	gitCommand := parts[0]
	if !slices.Contains(allowedCommands, gitCommand) {
		_, _ = fmt.Fprintf(session.Stderr(), "command not supported: %q\n", command)
		return
	}

	gitServicePack := strings.TrimPrefix(gitCommand, "git-")
	service, err := constant.ParseGitServiceType(gitServicePack)
	if err != nil {
		_, _ = fmt.Fprintf(session.Stderr(), "failed to parse service pack: %q\n", gitServicePack)
		return
	}

	// git command args
	gitArgs := parts[1:]

	// first git service pack cmd arg is path: 'space/repository.git' so we need to remove
	// single quotes.
	repoRef := strings.Trim(gitArgs[0], "'")
	// remove .git suffix
	repoRef = strings.TrimSuffix(repoRef, ".git")

	gitProtocol := ""
	for _, key := range session.Environ() {
		if strings.HasPrefix(key, "GIT_PROTOCOL=") {
			gitProtocol = key[len("GIT_PROTOCOL="):]
		}
	}

	ctx, cancel := context.WithCancel(session.Context())
	defer cancel()

	// set keep alive connection
	if s.KeepAliveInterval > 0 {
		go sendKeepAliveMsg(ctx, session, s.KeepAliveInterval)
	}

	err = s.RepoCtrl.GitServicePack(
		WithContext(ctx, session),
		&auth.Session{
			Principal: principal.User{
				ID:    principalInfo.ID,
				UID:   principalInfo.UID,
				Email: principalInfo.Email,
			},
		},
		repoRef,
		api.ServicePackOptions{
			Service:  service,
			Stdout:   session,
			Stdin:    session,
			Stderr:   session.Stderr(),
			Protocol: gitProtocol,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("git service pack failed")
		_, err = io.Copy(session.Stderr(), strings.NewReader(err.Error()))
		if err != nil {
			log.Error().Err(err).Msg("error writing to session stderr")
		}
	}
}

// - 定期向 SSH 客户端发送保活消息
// - 防止长时间运行的 Git 操作时连接超时
func sendKeepAliveMsg(ctx context.Context, session ssh.Session, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Ctx(ctx).Debug().Str("remote_addr", session.RemoteAddr().String()).Msgf("sendKeepAliveMsg")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Ctx(ctx).Debug().Msg("send keepalive message to ssh client")
			_, err := session.SendRequest(KeepAliveMsg, true, nil)
			if err != nil {
				log.Ctx(ctx).Debug().Err(err).Msg("failed to send keepalive message to ssh client")
			}
		}
	}
}

// - SSH 连接失败时的回调函数
// - 记录连接失败的日志
func sshConnectionFailed(conn net.Conn, err error) {
	log.Err(err).Msgf("failed connection from %s with error: %v", conn.RemoteAddr(), err)
}
