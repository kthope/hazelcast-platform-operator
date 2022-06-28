package managementcenter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-platform-operator/api/v1alpha1"
	n "github.com/hazelcast/hazelcast-platform-operator/internal/naming"
)

// Section contains the REST API endpoints.
const (
	createToken       = "/api/tokens"
	clusterConnection = "/rest/mc/clusterConnectionConfigs"
)

type HzClusters struct {
	ClusterName     string   `json:"clusterName,omitempty"`
	MemberAddresses []string `json:"memberAddresses,omitempty"`
}

type RestClient struct {
	url     string
	mcAdmin *McAdmin
}

func NewRestClient(mc *v1alpha1.ManagementCenter, mcAdmin *McAdmin) *RestClient {
	return &RestClient{
		url:     restUrl(mc),
		mcAdmin: mcAdmin,
	}
}

func (c *RestClient) CreateToken(ctx context.Context) (string, error) {
	reqBody, err := json.Marshal(c.mcAdmin)
	if err != nil {
		return "", err
	}
	ctxT, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := request(ctxT, "POST", reqBody, c.url, createToken, c.basicAuth)
	if err != nil {
		return "", err
	}
	res, err := c.executeRequest(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	admin := &McAdmin{}
	err = json.Unmarshal(b, admin)
	if err != nil {
		return "", err
	}
	c.mcAdmin.Token = admin.Token
	return admin.Token, nil
}

func (c *RestClient) GetHzClusters(ctx context.Context) ([]*HzClusters, error) {
	ctxT, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := request(ctxT, "GET", nil, c.url, clusterConnection, c.bearerAuth)
	if err != nil {
		return nil, err
	}
	res, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resp []*HzClusters
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %s", err)
	}

	if err := json.Unmarshal(resBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response's body error: %s", err)
	}
	return resp, nil
}

func (c *RestClient) CreateUpdateHzCluster(ctx context.Context, hzCluster *HzClusters) error {
	reqBody, err := json.Marshal(hzCluster)
	if err != nil {
		return err
	}
	ctxT, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := request(ctxT, "POST", reqBody, c.url, clusterConnection, c.bearerAuth)
	if err != nil {
		return err
	}
	res, err := c.executeRequest(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (c *RestClient) DeleteHzCluster(ctx context.Context, clusterName string) error {
	parm := url.Values{}
	parm.Add("clusterName", clusterName)

	ctxT, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := request(ctxT, "DELETE", []byte(parm.Encode()), c.url, clusterConnection, c.bearerAuth)
	if err != nil {
		return err
	}
	res, err := c.executeRequest(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (c *RestClient) executeRequest(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, res.Body)
		return res, fmt.Errorf("unexpected HTTP error: %s, %s", res.Status, buf.String())
	}
	return res, nil
}

func request(ctx context.Context, method string, data []byte, url string, endpoint string, auth func(*http.Request)) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url+endpoint, bytes.NewBuffer(data))
	req.Header.Add("Content-Type", "application/json")
	auth(req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *RestClient) basicAuth(req *http.Request) {
	req.SetBasicAuth(c.mcAdmin.Username, c.mcAdmin.Password)
}

func (c *RestClient) bearerAuth(req *http.Request) {
	var bearer = "Bearer " + c.mcAdmin.Token
	req.Header.Add("Authorization", bearer)
}

func restUrl(mc *v1alpha1.ManagementCenter) string {
	return fmt.Sprintf("http://%s", managementCenterURL(mc))
}

func managementCenterURL(mc *v1alpha1.ManagementCenter) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", mc.Name, mc.Namespace, n.DefaultMcPort)
}
