package commands

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"

	"github.com/Dreamacro/clash-ctl/common"
	"github.com/Dreamacro/clash-ctl/utils"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func HandleProxyCommand(args []string) {
	proxiesM, proxies, err := GetProxiesHash()
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
		if err != nil {
			fmt.Println(text.FgRed.Sprint(err.Error()))
		}
		params := args[1:]
		switch len(params) {
		case 0:
			t := table.NewWriter()
			t.SetStyle(table.StyleRounded)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Alias", "Name", "Type", "Now"})
			rows := []table.Row{}

			for i, p := range proxiesM {
				if p.Type == "Selector" {
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
				fmt.Println(text.FgRed.Sprint("Can't find proxy: " + params[0]))
				return
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
					alias, _ := getAliasWithConfig(node.Name)
					rows = append(rows, table.Row{
						alias,
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

		group := proxiesM[args[1]].Name
		groupEscaped := url.PathEscape(proxiesM[args[1]].Name)
		proxy := proxiesM[args[2]].Name

		body := map[string]interface{}{
			"name": proxy,
		}
		fail := common.HTTPError{}

		resp, err := req.R().SetError(&fail).SetBody(body).Put("/proxies/" + groupEscaped)
		if err != nil {
			fmt.Println(text.FgRed.Sprint(err.Error()))
			return
		}

		if resp.IsError() {
			fmt.Println(text.FgRed.Sprint(fail.Message))
		} else {
			fmt.Println(text.FgGreen.Sprint("Set proxy " + proxy + " to group " + group))
		}
	case "delay":
		wg := sync.WaitGroup{}
		for _, proxy := range proxies {
			go speedTest(proxy, server, &wg)
		}
		wg.Wait()
	}
}

func speedTest(proxy Proxy, server *common.Server, wg *sync.WaitGroup) {
	wg.Add(1)
	if proxy.Type == "Vmess" || proxy.Type == "ShadowsocksR" {
		nameEscaped := url.PathEscape(proxy.Name)
		req := common.MakeRequest(*server)
		fail := common.HTTPError{}

		result := struct {
			Delay int `json:"delay"`
		}{}

		resp, err := req.R().SetError(&fail).SetResult(&result).SetQueryParams(map[string]string{
			"timeout": "5000",
			"url":     "http://www.gstatic.com/generate_204",
		}).Get("/proxies/" + nameEscaped + "/delay")
		if err != nil {
			fmt.Println(text.FgRed.Sprint(err.Error()))
			return
		}
		alias, _ := getAliasWithConfig(proxy.Name)
		if resp.IsError() {
			fmt.Println(text.FgRed.Sprintf("%s \t %s \t %s", alias, proxy.Name, fail.Message))
		} else {
			fmt.Println(text.FgGreen.Sprintf("%s \t %s \t %d ms", alias, proxy.Name, result.Delay))
		}
	}
	wg.Done()
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
		// TODO: refactor duplicate code
		proxiesM, _, err := GetProxiesHash()
		if err != nil {
			return 0, nodes
		}
		for hashed, proxy := range proxiesM {
			if proxy.Type == "Selector" {
				nodes = append(nodes, common.Node{
					Text:        strings.Replace(hashed, " ", "%20", -1),
					Description: fmt.Sprintf("%s select `%s` now", proxy.Name, proxy.Now),
				})
			}
		}
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Description < nodes[j].Description })
	return len(params), nodes
}

func ProxySetResolver(params []string) (int, []common.Node) {
	nodes := []common.Node{}
	proxiesA, proxies, err := GetProxiesHash()

	switch len(params) {
	case 1:
		// TODO: refactor duplicate code
		if err != nil {
			return 0, nodes
		}
		for hashed, proxy := range proxiesA {
			if proxy.Type == "Selector" {
				nodes = append(nodes, common.Node{
					Text:        hashed,
					Description: fmt.Sprintf("%s select `%s` now", proxy.Name, proxy.Now),
				})
			}
		}
	case 2:
		groupName := proxiesA[params[0]].Name
		realName := strings.Replace(groupName, "%20", " ", -1)
		group, err := GetProxyGroup(realName)
		if err != nil {
			return 0, nodes
		}

		for _, proxyName := range group.All {
			proxy := proxies[proxyName]
			delay := 0
			if len(proxy.History) > 0 {
				delay = proxy.History[len(proxy.History)-1].Delay
			}
			alias, _ := getAliasWithConfig(proxy.Name)
			nodes = append(nodes, common.Node{
				Text:        alias,
				Description: fmt.Sprintf("%s %d ms", proxy.Name, delay),
			})
		}
	}

	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Description < nodes[j].Description })
	return len(params), nodes
}

func getAlias(proxyName string,
	replace map[string]common.Replace,
	regex map[string]common.ReplaceRegex) string {
	var aliasList []string
	// The excecution order of serach and replace is random
	for _, r := range replace {
		for _, str := range r.From {
			if strings.Contains(proxyName, str) {
				aliasList = append(aliasList, r.To)
				break
			}
		}
	}
	// The excecution order of regex is random
	// use priority order to fix that
	for _, rreg := range regex {
		r := regexp2.MustCompile(rreg.Pattern, 0)
		if match, _ := r.FindStringMatch(proxyName); match != nil {
			str := match.String()
			final := strings.Replace(rreg.To, "$&", str, -1)
			if final != rreg.To {
				aliasList = append(aliasList, final)
			}
		}
	}
	hashed := utils.GenHashString(proxyName)[:4]
	aliasList = append(aliasList, hashed)
	return strings.Join(aliasList, "-")
}

// If can't read the config file, fallback to hashed name
func getAliasWithConfig(proxyName string) (string, error) {
	cfg, err := common.ReadCfg()
	if err != nil {
		return utils.GenHashString(proxyName)[:4], err
	}
	return getAlias(proxyName, cfg.Replace, cfg.Regex), nil
}

func GetProxiesHash() (map[string]Proxy, map[string]Proxy, error) {
	proxies, err := GetProxies()
	proxiesMap := map[string]Proxy{}
	if err != nil {
		return nil, nil, err
	}
	for _, p := range proxies {
		alias, _ := getAliasWithConfig(p.Name)
		proxiesMap[alias] = p
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
