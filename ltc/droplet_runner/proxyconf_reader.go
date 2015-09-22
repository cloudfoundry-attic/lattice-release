package droplet_runner

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type HTTPProxyConfReader struct {
	URL string
}

func (p *HTTPProxyConfReader) ProxyConf() (ProxyConf, error) {
	resp, err := http.Get(p.URL)
	if err != nil {
		return ProxyConf{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return ProxyConf{}, nil
	}

	if resp.StatusCode != 200 {
		return ProxyConf{}, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ProxyConf{}, err
	}

	proxyConf := ProxyConf{}
	if err := json.Unmarshal(body, &proxyConf); err != nil {
		return ProxyConf{}, err
	}

	return proxyConf, nil
}
