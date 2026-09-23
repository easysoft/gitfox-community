// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package url

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/url_provider.go -package=mock
import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	// GITSuffix is the suffix used to terminate repo paths for git apis.
	GITSuffix = ".git"

	// APIMount is the prefix path for the api endpoints.
	APIMount = "api"

	// GITMount is the prefix path for the git endpoints.
	GITMount = "git"

	// ArtifactMount is the prefix path for the artifact endpoints
	ArtifactMount = "_artifacts"

	// ContainerRegistry is the prefix path for the container registry endpoints
	ContainerRegistry = "v2"
)

// Provider is an abstraction of a component that provides system related URLs.
// NOTE: Abstract to allow for custom implementation for more complex routing environments.
type Provider interface {
	// GetInternalAPIURL returns the internally reachable base url of the server.
	// NOTE: url is guaranteed to not have any trailing '/'.
	GetInternalAPIURL(ctx context.Context) string

	// GenerateContainerGITCloneURL generates a URL that can be used by CI container builds to
	// interact with Harness and clone a repo.
	GenerateContainerGITCloneURL(ctx context.Context, repoPath string) string

	// GenerateGITCloneURL generates the public git clone URL for the provided repo path.
	// NOTE: url is guaranteed to not have any trailing '/'.
	GenerateGITCloneURL(ctx context.Context, repoPath string) string

	// GenerateCustomUIRepoURL returns the url for the UI screen of a repository.
	GenerateCustomUIRepoURL(repoPath, custom string) string
	// GenerateGITCloneSSHURL generates the public git clone URL for the provided repo path.
	// NOTE: url is guaranteed to not have any trailing '/'.
	GenerateGITCloneSSHURL(ctx context.Context, repoPath string) string

	// GenerateMirrorGITCloneURL generates the HTTP clone URL for a mirror repository.
	// It uses the external provider URL and connector slug instead of the internal repo path.
	GenerateMirrorGITCloneURL(ctx context.Context, providerURL, slug string) string

	// GenerateMirrorGITCloneSSHURL generates the SSH clone URL for a mirror repository.
	// It uses the external provider URL and connector slug instead of the internal repo path.
	GenerateMirrorGITCloneSSHURL(ctx context.Context, providerURL, slug string) string

	// GenerateMirrorSVNCloneURL generates the HTTP clone URL for a mirror SVN repository.
	// SVN clone URLs do not use the .git suffix or /git/ mount path.
	GenerateMirrorSVNCloneURL(ctx context.Context, providerURL, slug string) string

	// GenerateUIRepoURL returns the url for the UI screen of a repository.
	GenerateUIRepoURL(ctx context.Context, repoPath string) string

	// GenerateUIPRURL returns the url for the UI screen of an existing pr.
	GenerateUIPRURL(ctx context.Context, repoPath string, prID int64) string

	// GenerateUICompareURL returns the url for the UI screen comparing two references.
	GenerateUICompareURL(ctx context.Context, repoPath string, ref1 string, ref2 string) string

	// GetAPIHostname returns the host for the api endpoint.
	GetAPIHostname(ctx context.Context) string

	// GenerateUIBuildURL returns the endpoint to use for viewing build executions.
	GenerateUIBuildURL(ctx context.Context, repoPath, pipelineIdentifier string, seqNumber int64) string

	// GetGITHostname returns the host for the git endpoint.
	GetGITHostname(ctx context.Context) string

	// GetAPIProto returns the proto for the API hostname
	GetAPIProto(ctx context.Context) string

	// GenerateRegistryURL return the container registry endpoint
	GenerateRegistryURL(segments ...string) *url.URL
	// RegistryURL returns the url for oci token endpoint
	RegistryURL() string

	GenerateArtifactURL(segments ...string) *url.URL
}

// Provider provides the URLs of the Harness system.
type provider struct {
	// internalURL stores the URL via which the service is reachable at internally
	// (no need for internal services to go via public route).
	internalURL *url.URL

	// containerURL stores the URL that can be used to communicate with Harness from inside a
	// build container.
	containerURL *url.URL

	// apiURL stores the raw URL the api endpoints are reachable at publicly.
	apiURL *url.URL

	// gitURL stores the URL the git endpoints are available at.
	// NOTE: we store it as url.URL so we can derive clone URLS without errors.
	gitURL *url.URL

	SSHEnabled     bool
	SSHDefaultUser string
	gitSSHURL      *url.URL

	// uiURL stores the raw URL to the ui endpoints.
	uiURL *url.URL

	// registryURL stores the raw URL to the registry endpoints.
	registryURL *url.URL
}

func NewProvider(
	internalURLRaw,
	containerURLRaw string,
	apiURLRaw string,
	gitURLRaw,
	gitSSHURLRaw string,
	sshDefaultUser string,
	sshEnabled bool,
	uiURLRaw string,
	registryURLRaw string,
) (Provider, error) {
	// remove trailing '/' to make usage easier
	internalURLRaw = strings.TrimRight(internalURLRaw, "/")
	containerURLRaw = strings.TrimRight(containerURLRaw, "/")
	apiURLRaw = strings.TrimRight(apiURLRaw, "/")
	gitURLRaw = strings.TrimRight(gitURLRaw, "/")
	gitSSHURLRaw = strings.TrimRight(gitSSHURLRaw, "/")
	uiURLRaw = strings.TrimRight(uiURLRaw, "/")
	registryURLRaw = strings.TrimRight(registryURLRaw, "/")

	internalURL, err := url.Parse(internalURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided internalURLRaw '%s' is invalid: %w", internalURLRaw, err)
	}

	containerURL, err := url.Parse(containerURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided containerURLRaw '%s' is invalid: %w", containerURLRaw, err)
	}

	apiURL, err := url.Parse(apiURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided apiURLRaw '%s' is invalid: %w", apiURLRaw, err)
	}

	gitURL, err := url.Parse(gitURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided gitURLRaw '%s' is invalid: %w", gitURLRaw, err)
	}

	gitSSHURL, err := url.Parse(gitSSHURLRaw)
	if sshEnabled && err != nil {
		return nil, fmt.Errorf("provided gitSSHURLRaw '%s' is invalid: %w", gitSSHURLRaw, err)
	}

	uiURL, err := url.Parse(uiURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided uiURLRaw '%s' is invalid: %w", uiURLRaw, err)
	}

	registryURL, err := url.Parse(registryURLRaw)
	if err != nil {
		return nil, fmt.Errorf("provided registryURLRaw '%s' is invalid: %w", registryURLRaw, err)
	}

	return &provider{
		internalURL:    internalURL,
		containerURL:   containerURL,
		apiURL:         apiURL,
		gitURL:         gitURL,
		gitSSHURL:      gitSSHURL,
		SSHDefaultUser: sshDefaultUser,
		SSHEnabled:     sshEnabled,
		uiURL:          uiURL,
		registryURL:    registryURL,
	}, nil
}

func (p *provider) GetInternalAPIURL(context.Context) string {
	return p.internalURL.JoinPath(APIMount).String()
}

func (p *provider) GenerateContainerGITCloneURL(_ context.Context, repoPath string) string {
	repoPath = path.Clean(repoPath)
	if !strings.HasSuffix(repoPath, GITSuffix) {
		repoPath += GITSuffix
	}

	return p.containerURL.JoinPath(GITMount, repoPath).String()
}

func (p *provider) GenerateGITCloneURL(_ context.Context, repoPath string) string {
	repoPath = path.Clean(repoPath)
	if !strings.HasSuffix(repoPath, GITSuffix) {
		repoPath += GITSuffix
	}

	return p.gitURL.JoinPath(repoPath).String()
}

func (p *provider) GenerateGITCloneSSHURL(_ context.Context, repoPath string) string {
	if !p.SSHEnabled {
		return ""
	}
	return BuildGITCloneSSHURL(p.SSHDefaultUser, p.gitSSHURL, repoPath)
}

func (p *provider) GenerateMirrorGITCloneURL(_ context.Context, providerURL, slug string) string {
	u, err := url.Parse(providerURL)
	if err != nil {
		return ""
	}

	slug = path.Clean(slug)
	if !strings.HasSuffix(slug, GITSuffix) {
		slug += GITSuffix
	}

	return u.JoinPath(slug).String()
}

func (p *provider) GenerateMirrorGITCloneSSHURL(_ context.Context, providerURL, slug string) string {
	u, err := url.Parse(providerURL)
	if err != nil {
		return ""
	}

	slug = path.Clean(slug)
	if !strings.HasSuffix(slug, GITSuffix) {
		slug += GITSuffix
	}

	host := u.Hostname()
	port := u.Port()
	combinedPath := strings.TrimLeft(slug, "/")

	if port != "" && port != "0" && port != "22" {
		return fmt.Sprintf(
			"ssh://git@%s/%s",
			net.JoinHostPort(host, port), combinedPath,
		)
	}

	return fmt.Sprintf(
		"git@%s:%s",
		host, combinedPath,
	)
}

func (p *provider) GenerateMirrorSVNCloneURL(_ context.Context, providerURL, slug string) string {
	u, err := url.Parse(providerURL)
	if err != nil {
		return ""
	}

	// Normalize "scheme:///path" → "scheme://host/path".
	// url.Parse puts everything after :// into Path when no host is present,
	// but SVN URLs should always have a host (the first path segment).
	if u.Host == "" && u.Path != "" {
		trimmed := strings.Trim(u.Path, "/")
		if idx := strings.Index(trimmed, "/"); idx > 0 {
			u.Host = trimmed[:idx]
			u.Path = "/" + trimmed[idx+1:]
		} else {
			u.Host = trimmed
			u.Path = ""
		}
	}

	trimmedSlug := strings.Trim(slug, "/")

	// Dedup: slug = "host/path" or "host" — providerURL already fully contains slug.
	if u.Host != "" {
		hostAndPath := u.Host
		if p := strings.Trim(u.Path, "/"); p != "" {
			hostAndPath += "/" + p
		}
		if hostAndPath == trimmedSlug {
			return u.String()
		}
	}

	// Dedup: slug = path only — providerURL's path already matches.
	if trimmedPath := strings.Trim(u.Path, "/"); trimmedPath != "" && trimmedPath == trimmedSlug {
		return u.String()
	}

	// Build: no host in providerURL — the slug provides the host.
	if u.Host == "" {
		if idx := strings.Index(trimmedSlug, "/"); idx > 0 {
			return u.Scheme + "://" + trimmedSlug[:idx] + "/" + trimmedSlug[idx+1:]
		}
		return u.Scheme + "://" + trimmedSlug
	}

	return u.JoinPath(slug).String()
}

func (p *provider) GenerateUIBuildURL(_ context.Context, repoPath, pipelineIdentifier string, seqNumber int64) string {
	return p.uiURL.JoinPath(
		repoPath, "pipelines",
		pipelineIdentifier, "execution", strconv.Itoa(int(seqNumber)),
	).String()
}

func (p *provider) GenerateUIRepoURL(_ context.Context, repoPath string) string {
	return p.uiURL.JoinPath(repoPath).String()
}

func (p *provider) GenerateCustomUIRepoURL(repoPath, custom string) string {
	return p.uiURL.JoinPath(repoPath, custom).String()
}

func (p *provider) GenerateUIPRURL(_ context.Context, repoPath string, prID int64) string {
	return p.uiURL.JoinPath(repoPath, "pulls", fmt.Sprint(prID)).String()
}

func (p *provider) GenerateUICompareURL(_ context.Context, repoPath string, ref1 string, ref2 string) string {
	return p.uiURL.JoinPath(repoPath, "pulls/compare", ref1+"..."+ref2).String()
}

func (p *provider) GetAPIHostname(context.Context) string {
	return p.apiURL.Hostname()
}

func (p *provider) GetGITHostname(context.Context) string {
	return p.gitURL.Hostname()
}

func (p *provider) GetAPIProto(context.Context) string {
	return p.apiURL.Scheme
}

func (p *provider) GenerateRegistryURL(segments ...string) *url.URL {
	l := []string{ContainerRegistry}
	l = append(l, segments...)
	return p.uiURL.JoinPath(l...)
}

func (p *provider) RegistryURL() string {
	return p.registryURL.String()
}

func (p *provider) GenerateArtifactURL(segments ...string) *url.URL {
	l := []string{ArtifactMount}
	l = append(l, segments...)
	return p.uiURL.JoinPath(l...)
}

func BuildGITCloneSSHURL(user string, sshURL *url.URL, repoPath string) string {
	repoPath = path.Clean(repoPath)
	if !strings.HasSuffix(repoPath, GITSuffix) {
		repoPath += GITSuffix
	}

	// SSH clone url requires custom format depending on port to satisfy git
	combinedPath := strings.Trim(path.Join(sshURL.Path, repoPath), "/")

	// handle custom ports differently as otherwise git clone fails
	if sshURL.Port() != "" && sshURL.Port() != "0" && sshURL.Port() != "22" {
		return fmt.Sprintf(
			"ssh://%s@%s/%s",
			user, net.JoinHostPort(sshURL.Hostname(), sshURL.Port()), combinedPath,
		)
	}

	return fmt.Sprintf(
		"%s@%s:%s",
		user, sshURL.Hostname(), combinedPath,
	)
}
