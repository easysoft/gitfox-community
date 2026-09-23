// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package webhook

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

var (
	errLoopbackNotAllowed       = errors.New("loopback not allowed")
	errPrivateNetworkNotAllowed = errors.New("private network not allowed")
)

func newHTTPClient(allowLoopback bool, allowPrivateNetwork bool, disableSSLVerification bool) *http.Client {
	// no customizations? use default client
	if allowLoopback && allowPrivateNetwork && !disableSSLVerification {
		return http.DefaultClient
	}

	// Clone http.DefaultTransport (used by http.DefaultClient)
	tr := http.DefaultTransport.(*http.Transport).Clone()

	tr.TLSClientConfig.InsecureSkipVerify = disableSSLVerification

	// create basic net.Dialer (Similar to what is used by http.DefaultTransport)
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// overwrite DialContext method to block sending data to localhost
	// NOTE: this doesn't block establishing the connection, but closes it before data is send.
	// WARNING: this allows scanning of IP addresses based on error types.
	tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		// dial connection using
		con, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		// by default close connection unless explicitly marked to keep it
		keepConnection := false
		defer func() {
			// if we decided to keep the connection, nothing to do
			if keepConnection {
				return
			}

			// otherwise best effort close connection
			cErr := con.Close()
			if cErr != nil {
				logger.GetLogger(ctx).Warnw(fmt.Sprintf("failed to close potentially malicious connection to '%s' (resolved: '%s')",
					addr, con.RemoteAddr()), "error", err)
			}
		}()

		// ensure a tcp address got established and close if it's localhost or private
		tcpAddr, ok := con.RemoteAddr().(*net.TCPAddr)
		if !ok {
			// not expected to happen, but to be sure
			return nil, fmt.Errorf("address resolved to a non-TCP address (original: '%s', resolved: '%s')",
				addr, con.RemoteAddr())
		}

		if !allowLoopback && tcpAddr.IP.IsLoopback() {
			return nil, errLoopbackNotAllowed
		}

		if !allowPrivateNetwork && tcpAddr.IP.IsPrivate() {
			return nil, errPrivateNetworkNotAllowed
		}

		// otherwise keep connection
		keepConnection = true

		return con, nil
	}

	// httpClient is similar to http.DefaultClient, just with custom http.Transport
	return &http.Client{Transport: tr}
}
