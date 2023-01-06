package apollo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Config struct {
	AppId       string
	ServerUrl   string
	ClusterName string
	Namespaces  []string
	release     map[string]string
	Timeout     time.Duration
}

type ConfigRes struct {
	AppId          string                 `json:"appId"`
	Cluster        string                 `json:"cluster"`
	NamespaceName  string                 `json:"namespaceName"`
	Configurations map[string]interface{} `json:"configurations"`
	ReleaseKey     string                 `json:"releaseKey"`
}

func GetConfig(c Config) (map[string]map[string]interface{}, error) {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
		},
		Timeout: c.Timeout,
	}
	conf := make(map[string]map[string]interface{})
	for _, v := range c.Namespaces {
		u := fmt.Sprintf("%s/configs/%s/%s/%s", c.ServerUrl, c.AppId, c.ClusterName, v)
		resp, err := client.Get(u)
		if err != nil {
			return conf, err
		}
		if resp.StatusCode != http.StatusOK {
			return conf, errors.New("not ok status")
		}
		defer resp.Body.Close()
		r, _ := ioutil.ReadAll(resp.Body)
		confRes := ConfigRes{}
		json.Unmarshal(r, &confRes)
		conf[confRes.NamespaceName] = confRes.Configurations
	}
	return conf, nil
}
