package commands

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/Dreamacro/clash-ctl/common"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func genSha1String(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

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
	case "ls":
		proxiesM, proxies, err := GetProxiesSha1()
		proxiesList := []Proxy{}

		if err != nil {
			fmt.Println(text.FgRed.Sprint(err.Error()))
		}
		params := args[1:]
		switch len(params) {
		case 0:
			t := table.NewWriter()
			t.SetStyle(table.StyleRounded)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Index", "Name", "Type", "Now"})
			rows := []table.Row{}

			for i, p := range proxiesM {
				if p.Type == "Selector" {
					proxiesList = append(proxiesList, p)
					rows = append(rows, table.Row{
						i,
						p.Name,
						p.Type,
						p.Now,
					})
				}
			}

			t.AppendRows(rows)
			t.Render()
		case 1:
			t := table.NewWriter()
			t.SetStyle(table.StyleRounded)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Index", "Name", "Type", "Delay"})
			rows := []table.Row{}

			val, ok := proxiesM[params[0]]
			if !ok {
				// no shadows. override
				val = proxies[params[0]]
				if !ok {
					fmt.Println(text.FgRed.Sprint("Can't find proxy: " + params[0]))
					return
				}
			}
			if val.Type != "Selector" {
				fmt.Println(text.FgRed.Sprint("Please select a Selector type instead of nodes"))
				return
			}
			for _, name := range val.All {
				node, ok := proxies[name]
				delay := 0
				if len(node.History) > 0 {
					delay = node.History[len(node.History)-1].Delay
				}
				if ok {
					rows = append(rows, table.Row{
						genSha1String(node.Name)[:4],
						node.Name,
						node.Type,
						delay,
					})
				}
			}

			t.AppendRows(rows)
			t.Render()
		}

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
		Time  string `json:"time"` // time.Time RFC3339
		Delay int    `json:"delay"`
	} `json:"history"`
}

func ProxyListResolver(params []string) (int, []common.Node) {
	nodes := []common.Node{}

	switch len(params) {
	case 1:
		proxiesM, _, err := GetProxiesSha1()
		if err != nil {
			return 0, nodes
		}
		for sha1, proxy := range proxiesM {
			if proxy.Type == "Selector" {
				nodes = append(nodes, common.Node{
					Text:        strings.Replace(sha1, " ", "%20", -1),
					Description: fmt.Sprintf("select `%s` now", proxy.Name),
				})
			}
		}
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Description < nodes[j].Description })
	return len(params), nodes
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

func GetProxiesSha1() (map[string]Proxy, map[string]Proxy, error) {
	proxies, err := GetProxies()
	proxiesMap := map[string]Proxy{}
	if err != nil {
		return nil, nil, err
	}
	for _, p := range proxies {
		proxiesMap[genSha1String(p.Name)[:4]] = p
	}

	return proxiesMap, proxies, nil
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
