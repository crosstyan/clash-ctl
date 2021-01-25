package commands

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Dreamacro/clash-ctl/common"

	"github.com/jedib0t/go-pretty/v6/text"
)

func HandleProxyCommand(args []string) {
	if len(args) == 0 {
		return
	}

	cfg, err := common.ReadCfg()
	if err != nil {
		return
	}

	_, server, err := common.GetCurrentServer(cfg)
	if err != nil {
		fmt.Println(text.FgRed.Sprint(err.Error()))
		return
	}

	switch args[0] {
	case "set":
		req := common.MakeRequest(*server)
		if len(args) < 3 {
			fmt.Println(text.FgRed.Sprint("should be `set proxy group proxyName`"))
			return
		}

		group := url.PathEscape(strings.Replace(args[1], "%20", " ", -1))
		proxy := strings.Replace(args[2], "%20", " ", -1)

		body := map[string]interface{}{
			"name": proxy,
		}
		fail := common.HTTPError{}

		resp, err := req.R().SetError(&fail).SetBody(body).Put("/proxies/" + group)
		if err != nil {
			fmt.Println(text.FgRed.Sprint(err.Error()))
			return
		}

		if resp.IsError() {
			fmt.Println(text.FgRed.Sprint(fail.Message))
		}
	}
}

type Proxy struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Now     string   `json:"now"`
	All     []string `json:"all"`
	History []struct {
		Delay string `json:"delay"`
	} `json:"history"`
}

func ProxySetResolver(params []string) (int, []common.Node) {
	nodes := []common.Node{}

	switch len(params) {
	case 1:
		proxies, err := GetProxies()
		if err != nil {
			return 0, nodes
		}
		for name, proxy := range proxies {
			if proxy.Type == "Selector" {
				nodes = append(nodes, common.Node{
					Text:        strings.Replace(name, " ", "%20", -1),
					Description: fmt.Sprintf("select `%s` now", proxy.Now),
				})
			}
		}
	case 2:
		realName := strings.Replace(params[0], "%20", " ", -1)
		group, err := GetProxyGroup(realName)
		if err != nil {
			return 0, nodes
		}
		for _, proxy := range group.All {
			nodes = append(nodes, common.Node{
				Text: strings.Replace(proxy, " ", "%20", -1),
			})
		}
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Text < nodes[j].Text })
	return len(params), nodes
}

func GetProxies() (map[string]Proxy, error) {
	cfg, err := common.ReadCfg()
	if err != nil {
		return nil, err
	}

	_, server, err := common.GetCurrentServer(cfg)
	if err != nil {
		return nil, err
	}

	req := common.MakeRequest(*server)

	result := struct {
		Proxies map[string]Proxy `json:"proxies"`
	}{}
	_, err = req.R().SetResult(&result).Get("/proxies")
	if err != nil {
		return nil, err
	}

	return result.Proxies, nil
}

func GetProxyGroup(group string) (*Proxy, error) {
	cfg, err := common.ReadCfg()
	if err != nil {
		return nil, err
	}

	_, server, err := common.GetCurrentServer(cfg)
	if err != nil {
		return nil, err
	}

	req := common.MakeRequest(*server)

	result := &Proxy{}
	_, err = req.R().SetResult(result).Get("/proxies/" + url.PathEscape(group))
	if err != nil {
		return nil, err
	}

	return result, nil
}
