package metrics

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	utilerrors "github.com/appscode/go/util/errors"
	"github.com/pkg/errors"
	config_util "github.com/prometheus/common/config"
	prom_config "github.com/prometheus/common/config"
)

// xref: https://github.com/prometheus/prometheus/blob/master/storage/remote/client.go

const maxErrMsgLen = 256

var userAgent = "metrics-exporter"

// Client allows writing to a remote HTTP endpoint.
type RemoteClient struct {
	url     *config_util.URL
	client  *http.Client
	timeout time.Duration
	license string
}

func NewRemoteClient(addr string, license string, config prom_config.HTTPClientConfig, timeout time.Duration) (*RemoteClient, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	httpClient, err := config_util.NewClientFromConfig(config, "remote_storage")
	if err != nil {
		return nil, err
	}

	cl := &RemoteClient{
		url: &prom_config.URL{
			URL: u,
		},
		timeout: timeout,
		client:  httpClient,
		license: license,
	}
	return cl, nil
}

// Store sends a batch of samples to the HTTP endpoint, the request is the proto marshalled
// and encoded bytes from codec.go.
func (c *RemoteClient) Store(ctx context.Context, req []byte) (rerr error) {
	httpReq, rerr := http.NewRequest("POST", c.url.String(), bytes.NewReader(req))
	if rerr != nil {
		// Errors from NewRequest are from unparseable URLs, so are not
		// recoverable.
		return
	}
	httpReq.Header.Add("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("User-Agent", userAgent)
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	// Add authorization header
	if len(httpReq.Header.Get("Authorization")) == 0 {
		httpReq.Header.Set("Authorization", fmt.Sprintf("JWT %s", c.license))
	}

	httpReq = httpReq.WithContext(ctx)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	httpResp, rerr := c.client.Do(httpReq.WithContext(ctx))
	if rerr != nil {
		return
	}
	defer func() {
		_, e1 := io.Copy(ioutil.Discard, httpResp.Body)
		e2 := httpResp.Body.Close()
		rerr = utilerrors.NewAggregate([]error{rerr, e1, e2})
	}()

	if httpResp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(httpResp.Body, maxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		rerr = errors.Errorf("server returned HTTP status %s: %s", httpResp.Status, line)
	}
	return
}
