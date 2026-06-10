package importer

import (
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	userModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/paths"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/gitea"
	"github.com/drone/go-scm/scm/driver/github"
	"github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/driver/gogs"
	"github.com/drone/go-scm/scm/transport"
	"github.com/drone/go-scm/scm/transport/oauth2"
)

type RepositoryInfo struct {
	Space         string
	Identifier    string
	CloneURL      string
	IsPublic      bool
	DefaultBranch string
}

// ConvertToRepo converts the RepositoryInfo into the types.Repository object marked as being imported and is-public flag.
func (r *RepositoryInfo) ConvertToRepo(
	spaceID uint,
	identifier string,
	description string,
	acl string,
	mirror bool,
	product string,
	user *userModel.User,
	provider *repoModel.Provider,
) (*repoModel.Repo, bool) {
	var connectInfo []byte
	scmType := "git"
	if provider != nil {
		switch provider.Type {
		case constant.RepoProviderTypeSVN:
			svnConnectInfo := repoModel.SVNConnectInfo{
				Slug:     provider.Slug,
				User:     provider.Username,
				Password: base64.StdEncoding.EncodeToString([]byte(provider.Password)),
			}
			connectInfo, _ = json.Marshal(svnConnectInfo)
			scmType = "svn"
		default:
			gitConnectInfo := repoModel.GitConnectInfo{
				Slug:      provider.Slug,
				ProjectID: provider.ProjectID,
			}
			connectInfo, _ = json.Marshal(gitConnectInfo)
		}
	}

	now := time.Now().UnixMilli()
	gitTempUID := fmt.Sprintf("importing-%s-%d", hash(fmt.Sprintf("%d:%s", spaceID, identifier)), now)
	forkID := uint(0)
	return &repoModel.Repo{
		SpaceID:       spaceID,
		Identifier:    identifier,
		GitUID:        gitTempUID, // the correct git UID will be set by the job handler
		Desc:          description,
		SCMType:       scmType,
		CreatedBy:     user.Account,
		ForkID:        &forkID,
		DefaultBranch: r.DefaultBranch,
		Status:        constant.RepoStateImporting, // 标记为导入中状态
		ACL:           acl,
		Mirror:        mirror,
		Product:       product,
		Path:          paths.Concatenate(fmt.Sprint(spaceID), identifier),
		ProviderID:    provider.ID,
		Connector:     connectInfo,
	}, r.IsPublic
}

func hash(s string) string {
	h := sha512.New()
	_, _ = h.Write([]byte(s))
	return base32.StdEncoding.EncodeToString(h.Sum(nil)[:10])
}

func oauthTransport(token string, scheme string) http.RoundTripper {
	if token == "" {
		return nil
	}
	return &oauth2.Transport{
		Scheme: scheme,
		Source: oauth2.StaticTokenSource(&scm.Token{Token: token}),
	}
}

func authHeaderTransport(token string) http.RoundTripper {
	if token == "" {
		return nil
	}
	return &transport.Authorization{
		Scheme:      "token",
		Credentials: token,
	}
}

func basicAuthTransport(username, password string) http.RoundTripper {
	if username == "" && password == "" {
		return nil
	}
	return &transport.BasicAuth{
		Username: username,
		Password: password,
	}
}

// getScmClientWithTransport creates an SCM client along with the necessary transport
// layer depending on the provider. For example, for bitbucket we support app passwords
// so the auth transport is BasicAuth whereas it's Oauth for other providers.
// It validates that auth credentials are provided if authReq is true.
func getScmClientWithTransport(provider repoModel.Provider, slug string) (*scm.Client, error) { //nolint:gocognit

	if provider.Host == "" {
		return nil, errors.BadRequest.New("scm provider Host missing")
	}
	if provider.Token == "" {
		return nil, errors.BadRequest.Newf("scm provider authentication token missing for %s", provider.Type)
	}

	var c *scm.Client
	var err error
	var transport http.RoundTripper
	switch provider.Type {
	case "":
		return nil, errors.BadRequest.New("scm provider can not be empty")

	case constant.RepoProviderTypeGitHub:
		c, err = github.New(provider.Host)
		if err != nil {
			return nil, errors.BadRequest.Wrap(err, "scm provider Host invalid")
		}
		transport = oauthTransport(provider.Token, oauth2.SchemeBearer)

	case constant.RepoProviderTypeGitLab:
		c, err = gitlab.New(provider.Host)
		if err != nil {
			return nil, errors.BadRequest.Wrap(err, "scm provider Host invalid")
		}
		transport = oauthTransport(provider.Token, oauth2.SchemeBearer)

	case constant.RepoProviderTypeGitea:
		c, err = gitea.New(provider.Host)
		if err != nil {
			return nil, errors.BadRequest.Wrap(err, "scm provider Host invalid")
		}
		transport = authHeaderTransport(provider.Token)

	case constant.RepoProviderTypeGogs:
		c, err = gogs.New(provider.Host)
		if err != nil {
			return nil, errors.BadRequest.Wrap(err, "scm provider Host invalid")
		}
		transport = authHeaderTransport(provider.Token)

	default:
		return nil, errors.BadRequest.Newf("unsupported scm provider: %s", provider.Type)
	}

	if c == nil {
		return nil, errors.BadRequest.Newf("failed to initialize scm client for provider: %s", provider.Type)
	}

	// override default transport if available
	if transport != nil {
		c.Client = &http.Client{Transport: transport}
	}

	return c, nil
}

func extractOrgAndProjectFromSlug(slug string) (string, string, error) {
	res := strings.Split(slug, "/")
	if len(res) < 2 {
		return "", "", errors.BadRequest.New("organization or project info missing")
	}
	if len(res) > 3 {
		return "", "", errors.BadRequest.New("too many parts")
	}
	return res[0], res[1], nil
}

func extractRepoFromSlug(slug string) (string, error) {
	res := strings.Split(slug, "/")
	if len(res) == 3 {
		return res[2], nil
	}
	return "", errors.BadRequest.New("repo name missing")
}

func convertImportError(provider repoModel.Provider, slug string, r *scm.Response, err error) error {
	if err == nil {
		return nil
	}

	if r == nil {
		if provider.Host != "" {
			return errors.BadRequest.Newf("failed to make HTTP request to %s (host=%s): %s",
				provider.Type, provider.Host, err)
		}

		return errors.BadRequest.Newf("failed to make HTTP request to %s: %s",
			provider.Type, err)
	}

	switch r.Status {
	case http.StatusNotFound:
		return errors.BadRequest.Newf("couldn't find %s at %s: %s",
			slug, provider.Type, err.Error())
	case http.StatusUnauthorized:
		return errors.BadRequest.Newf("bad credentials provided for %s at %s: %s",
			slug, provider.Type, err.Error())
	case http.StatusForbidden:
		return errors.BadRequest.Newf("access denied to %s at %s: %s",
			slug, provider.Type, err.Error())
	default:
		return errors.BadRequest.Newf("failed to fetch %s from %s (HTTP status %d): %s",
			slug, provider.Type, r.Status, err.Error())
	}
}
