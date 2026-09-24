// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package rbac

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	requestUser   = "user"
	requestSpace  = "space"
	requestObject = "object"
)

var (
	WildcardIdentify          = &objectRequest{in: MatchWildcard}
	RecursiveWildcardIdentify = &objectRequest{in: MatchRecursiveWildcard}
)

type RequestIdentity interface {
	Kind() string
	Key() string
}

type userRequest struct {
	account string
}

func NewUser(account string) RequestIdentity {
	return &userRequest{account: account}
}

func (u *userRequest) Kind() string {
	return requestUser
}

func (u *userRequest) Key() string {
	return fmt.Sprintf("%s:%s", requestUser, u.account)
}

type spaceRequest struct {
	id int64
}

func NewSpace(id int64) RequestIdentity {
	return &spaceRequest{id: id}
}

func (s *spaceRequest) Kind() string {
	return requestSpace
}

func (s *spaceRequest) Key() string {
	return fmt.Sprintf("%s:%d", requestSpace, s.id)
}

// repoSpaceRequest 编码仓库级作用域：space:{spaceID}/repo:{repoID}
// 用于在 Casbin policy 的 space 字段中区分仓库级权限和空间级权限。
type repoSpaceRequest struct {
	spaceID int64
	repoID  int64
}

// NewRepoSpace 创建仓库级作用域标识。
// Key 格式为 "space:{spaceID}/repo:{repoID}"，支持三级权限继承链：
//
//	Repo (space:5/repo:42) → Space (space:5) → Global (space:0) → 系统默认
func NewRepoSpace(spaceID, repoID int64) RequestIdentity {
	return &repoSpaceRequest{spaceID: spaceID, repoID: repoID}
}

func (r *repoSpaceRequest) Kind() string {
	return requestSpace
}

func (r *repoSpaceRequest) Key() string {
	return fmt.Sprintf("%s:%d/repo:%d", requestSpace, r.spaceID, r.repoID)
}

type objectRequest struct {
	in string
}

func NewObject(in string) RequestIdentity {
	return &objectRequest{in: in}
}

func (o *objectRequest) Kind() string {
	return requestObject
}

func (o *objectRequest) Key() string {
	return o.in
}

// ParseIdentity 解析请求标识。
func ParseIdentity(s string) (RequestIdentity, error) {
	if s == MatchWildcard {
		return WildcardIdentify, nil
	}

	frames := strings.Split(s, ":")
	if len(frames) != 2 {
		return nil, fmt.Errorf("invalid identity: %s", s)
	}

	switch frames[0] {
	case requestUser:
		return NewUser(frames[1]), nil
	case requestSpace:
		pk, err := strconv.ParseInt(frames[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid identity: %s", s)
		}
		return NewSpace(pk), nil
	default:
		return nil, fmt.Errorf("invalid identity: %s", s)
	}
}
