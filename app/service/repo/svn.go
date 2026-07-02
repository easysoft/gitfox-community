package repo

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	svnModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/svn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	svnPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/svn"
)

// ExecSvn 是 SVN 子命令统一调度入口。
//
// 业务流程:
//  1. 校验 repo.SCMType == "svn" 与 req.Func 在白名单
//  2. 从 repo.Connector 解析 SVNConnectInfo,base64 解码 password
//  3. 按 req.Func 反序列化 args 到对应 pkg/svn *XxxParams,并调用 svnSvc 对应方法
//  4. 返回 svn 原生 stdout 与建议的 Content-Type
//
// contentType 由 service 决定（diff -> text/x-diff,其余 -> text/plain）,
// controller 仅按此设置响应头,不参与业务判断。
func (s *Service) ExecSvnCommand(ctx context.Context, repo repoModel.Repo, req svnModel.ReqForExec) (output string, contentType string, err error) {
	if repo.SCMType != constant.RepoScmTypeSvn {
		return "", "", errors.WithI18nKey(
			errors.Newf("repo %d scmType=%s is not svn", repo.ID, repo.SCMType),
			i18n.SvnNotSvnRepo,
		)
	}

	if !req.Command.IsValid() {
		return "", "", errors.WithI18nKey(
			errors.Newf("unsupported command: %s", req.Command),
			i18n.SvnInvalidCommand,
		)
	}

	username, password, err := parseSvnConnector(repo.Connector)
	if err != nil {
		return "", "", err
	}

	output, err = s.dispatchSvnCommand(ctx, req, repo.GitUID, username, password)
	if err != nil {
		// dispatchSvnCommand 内已包过 i18n 的(如 SvnArgsInvalid)不会被覆盖,其余底层错误统一标 SvnExecFailed。
		if errors.GetI18nKey(err) == i18n.ErrorDefault {
			err = errors.WithI18nKey(err, i18n.SvnExecFailed)
		}
		return "", "", err
	}

	return output, svnContentTypeOf(req.Command), nil
}

// dispatchSvnCommand 按 Func 把 Args 反序列化到对应 *XxxParams 并调 pkg/svn。
// svnUID 由 repo.GitUID 注入,与 pipeline 内部对 svnUID 的寻址语义一致；
// username/password 由 repo.Connector 中的 SVNConnectInfo 解出,不由调用方传入。
func (s *Service) dispatchSvnCommand(ctx context.Context, r svnModel.ReqForExec, svnUID, username, password string) (string, error) {
	rawArgs, _ := json.Marshal(r.Args)

	switch r.Command {
	case svnModel.Ls:
		p := &svnPkg.LsParams{}
		if err := json.Unmarshal(rawArgs, p); err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnInvalidArgs)
		}
		p.SvnUID, p.Username, p.Password = svnUID, username, password
		output, err := s.svnSvc.Ls(ctx, p)
		if err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnLsFailed)
		}
		return output, nil

	case svnModel.Log:
		p := &svnPkg.LogParams{}
		if err := json.Unmarshal(rawArgs, p); err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnInvalidArgs)
		}
		p.SvnUID, p.Username, p.Password = svnUID, username, password
		output, err := s.svnSvc.Log(ctx, p)
		if err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnLogFailed)
		}
		return output, nil

	case svnModel.Blame:
		p := &svnPkg.BlameParams{}
		if err := json.Unmarshal(rawArgs, p); err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnInvalidArgs)
		}
		p.SvnUID, p.Username, p.Password = svnUID, username, password
		output, err := s.svnSvc.Blame(ctx, p)
		if err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnBlameFailed)
		}
		return output, nil

	case svnModel.Cat:
		p := &svnPkg.CatParams{}
		if err := json.Unmarshal(rawArgs, p); err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnInvalidArgs)
		}
		p.SvnUID, p.Username, p.Password = svnUID, username, password
		output, err := s.svnSvc.Cat(ctx, p)
		if err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnCatFailed)
		}
		return output, nil

	case svnModel.Info:
		// Info 走 URL 而非 SvnUID(pkg/svn 当前签名如此),URL 由 svnUID 推导为 file://...；
		// Subpath 反序列化到 p.Subpath 后由 pkg/svn.Info 内部拼接。
		p := &svnPkg.InfoParams{}
		if err := json.Unmarshal(rawArgs, p); err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnInvalidArgs)
		}
		p.URL = s.svnSvc.FileURL(svnUID)
		p.Username, p.Password = username, password
		output, err := s.svnSvc.Info(ctx, p)
		if err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnInfoFailed)
		}
		return output, nil

	case svnModel.Diff:
		p := &svnPkg.DiffParams{}
		if err := json.Unmarshal(rawArgs, p); err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnInvalidArgs)
		}
		p.SvnUID, p.Username, p.Password = svnUID, username, password
		output, err := s.svnSvc.Diff(ctx, p)
		if err != nil {
			return "", errors.WithI18nKey(err, i18n.SvnDiffFailed)
		}
		return output, nil
	}

	// IsValid 已挡,此处理论不可达
	return "", errors.WithI18nKey(errors.New("unreachable"), i18n.SvnInvalidCommand)
}

// parseSvnConnector 从 repo.Connector 解析 SVN 凭证,与 mirror_sync 中的姿态保持一致：
// connector 是 JSON,形如 {"slug":"...","user":"...","password":"<base64>"}；password 需 base64 解码。
func parseSvnConnector(connector []byte) (string, string, error) {
	info := repoModel.SVNConnectInfo{}
	if err := json.Unmarshal(connector, &info); err != nil {
		return "", "", errors.WithI18nKey(err, i18n.SvnConnectorParseError)
	}
	if info.Password == "" {
		return info.User, "", nil
	}
	pwdBytes, err := base64.StdEncoding.DecodeString(info.Password)
	if err != nil {
		return "", "", errors.WithI18nKey(err, i18n.SvnPasswordDecodeError)
	}
	return info.User, string(pwdBytes), nil
}

// svnContentTypeOf 按 Func 推断响应 Content-Type。
// diff 给 text/x-diff,其余统一以 text/plain 流式返回；
// 调用方按方法语义自行 simplexml_load_string,不依赖 Content-Type 切换。
func svnContentTypeOf(f svnModel.Command) string {
	switch f {
	case svnModel.Diff:
		return "text/x-diff; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}
