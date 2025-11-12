package authn

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox/types/enum"
	"github.com/go-ldap/ldap/v3"
	"golang.org/x/crypto/bcrypt"
)

// BasicAuthenticator uses the provided JWT to authenticate the caller.
// BasicAuthenticator 使用提供的JWT来验证调用者。
type BasicAuthenticator struct {
	principalStore principal.IPrincipalStore
	config         *config.Config
	// log            *provider.SystemLog
	//ldap           *ldap.Conn
}

func (a *BasicAuthenticator) Authenticate(r *http.Request) (*auth.Session, error) {
	// a.log.Debug("[authn.basic]trigger BasicAuthenticator")
	ctx := r.Context()
	userName, passwd, ok := r.BasicAuth()
	if !ok {
		return nil, constant.ErrAuthNoData
	}

	user, err := a.principalStore.FindUserByUID(ctx, userName)
	if errors.Is(err, errors.ErrResourceNotFound) {
		return nil, constant.ErrUnauthorized
	}

	if user.Blocked {
		// a.log.Debugf("[authn.basic]user %s is blocked", user.UID)
		return nil, constant.ErrUnauthorized
	}

	switch user.Source {
	case "", enum.PrincipalSourceZentao:
		if user.Source == enum.PrincipalSourceZentao &&
			a.config.LDAP.Enable &&
			user.UID != a.config.Principal.Admin.UID {
			if err := a.ldapAuthenticate(&a.config.LDAP, userName, passwd); err != nil {
				return nil, err
			}
			break
		}

		if user.Source == enum.PrincipalSourceZentao {
			hash := md5.Sum([]byte(passwd))
			hashedPassword := hex.EncodeToString(hash[:])
			if hashedPassword != user.Password {
				// a.log.Debugf("[authn.basic]user %s has invalid zentao password", user.UID)
				return nil, constant.ErrUnauthorized
			}
		} else {
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwd)); err != nil {
				// a.log.Debugf("[authn.basic]user %s has invalid password", user.UID)
				return nil, constant.ErrUnauthorized
			}
		}

	case enum.PrincipalSourceLDAP:
		if a.config.LDAP.Enable {
			if err := a.ldapAuthenticate(&a.config.LDAP, userName, passwd); err != nil {
				return nil, err
			}
		} else {
			return nil, constant.ErrInternal
		}
	default:
		return nil, constant.ErrInternal
	}

	// a.log.Debugf("[authn.basic]user %s authenticated", user.UID)
	return &auth.Session{
		Principal: *user.ToPrincipal(),
		User:      user.ToPrincipal(),
	}, nil
}

func (a *BasicAuthenticator) ldapAuthenticate(conf *config.LDAPConfig, username, password string) error {
	// a.log.Debugf("Starting LDAP authentication ldap_url: %s, bind_dn: %s", conf.URL, conf.BindDN)

	l, err := ldap.DialURL(conf.URL)
	if err != nil {
		// a.log.Errorf("ldap connection failed: %v", err)
		return constant.ErrInternal
	}
	defer l.Close()

	// a.log.Debug("LDAP connection established, binding as admin")
	if err = l.Bind(conf.BindDN, conf.BindPassword); err != nil {
		// a.log.Errorf("The ldap administrator binding failed: %v", err)
		return constant.ErrInternal
	}

	filter := fmt.Sprintf(conf.UserFilter, ldap.EscapeFilter(username))
	// a.log.Debugf("LDAP search filter: %s, base_dn: %s, attributes: %s", filter, conf.BaseDN, conf.Attributes)

	attributes := strings.Split(conf.Attributes, ",")
	searchRequest := ldap.NewSearchRequest(
		conf.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1, 0, false,
		filter,
		attributes,
		nil,
	)

	// a.log.Debugf("LDAP search request: %s", searchRequest)
	sr, err := l.Search(searchRequest)
	if err != nil {
		// a.log.Errorf("LDAP search failed: %v", err)
		return constant.ErrBadRequest
	}
	// a.log.Debugf("LDAP search result entries count: %d", len(sr.Entries))
	if len(sr.Entries) != 1 {
		// a.log.Debugf("LDAP search did not return exactly one entry, count: %d", len(sr.Entries))
		return constant.ErrNotFound
	}

	entry := sr.Entries[0]
	// a.log.Debugf("LDAP entry found, dn: %s, attributes: %v", entry.DN, entry.Attributes)

	var userDNStr string
	if strings.EqualFold(conf.DNField, "DN") {
		userDNStr = entry.DN
		// a.log.Debugf("LDAP user DN resolved from DN field: %s", userDNStr)
	} else {
		attrVal := entry.GetAttributeValue(conf.DNField)
		if attrVal == "" {
			// a.log.Errorf("Failed to resolve user DN attribute for LDAP bind, dn_field: %s", conf.DNField)
			return constant.ErrInternal
		}
		if conf.DNTemplate != "" {
			userDNStr = fmt.Sprintf(conf.DNTemplate, attrVal)
			// a.log.Debugf("LDAP user DN constructed from template, dn_field: %s, dn_template: %s, user_dn: %s", conf.DNField, conf.DNTemplate, userDNStr)
		} else {
			// a.log.Errorf("DNTemplate not configured for non-DN DNField, dn_field: %s", conf.DNField)
			return constant.ErrInternal
		}
	}

	// a.log.Debugf("Attempting to bind as user, user_dn: %s", userDNStr)
	if err = l.Bind(userDNStr, password); err != nil {
		// a.log.Debugf("LDAP user bind failed, user_dn: %s, err: %v", userDNStr, err)
		return constant.ErrUserPass
	}

	// a.log.Debugf("LDAP authentication successful, user_dn: %s", userDNStr)
	return nil
}
