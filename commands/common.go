package commands

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/Dreamacro/clash-ctl/common"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func HandleCommonCommand(args []string) {
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
	case "traffic":
		conn, err := common.MakeWebsocket(*server, "/traffic")
		if err != nil {
			fmt.Println(text.FgRed.Sprint(err.Error()))
			return
		}

		body := struct {
			Upload   int64 `json:"up"`
			Download int64 `json:"down"`
		}{}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		for {
			select {
			case <-sigCh:
				signal.Stop(sigCh)
				fmt.Println()
				return
			default:
				if err := conn.ReadJSON(&body); err == nil {
					downText := text.AlignDefault.Apply(
						fmt.Sprintf("Download: %s", text.FgGreen.Sprint(progress.FormatBytes(body.Download))),
						18,
					)

					upText := text.AlignDefault.Apply(
						fmt.Sprintf("Upload: %s", text.FgGreen.Sprint(progress.FormatBytes(body.Upload))),
						16,
					)
					fmt.Printf("\033[2K\r%s %s", downText, upText)
				}
			}
		}
	case "connections":
		req := common.MakeRequest(*server)

		type metadata struct {
			NetWork string `json:"network"`
			Type    string `json:"type"`
			SrcIP   string `json:"sourceIP"`
			DstIP   string `json:"destinationIP"`
			SrcPort string `json:"sourcePort"`
			DstPort string `json:"destinationPort"`
			Host    string `json:"host"`
		}

		type tracker struct {
			UUID          string   `json:"id"`
			Metadata      metadata `json:"metadata"`
			UploadTotal   int64    `json:"upload"`
			DownloadTotal int64    `json:"download"`
			Start         string   `json:"start"`
			Chain         []string `json:"chains"`
			Rule          string   `json:"rule"`
			RulePayload   string   `json:"rulePayload"`
		}

		snapshot := struct {
			DownloadTotal int64     `json:"downloadTotal"`
			UploadTotal   int64     `json:"uploadTotal"`
			Connections   []tracker `json:"connections"`
		}{
			Connections: []tracker{},
		}

		_, err := req.R().SetResult(&snapshot).Get("/connections")
		if err != nil {
			fmt.Println(text.FgRed.Sprint(err.Error()))
			return
		}

		t := table.NewWriter()
		t.SetStyle(table.StyleRounded)
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Host", "Network", "Type", "Chain", "Rule", "Time"})

		rows := []table.Row{}

		sort.Slice(snapshot.Connections, func(i, j int) bool {
			l, _ := time.Parse(time.RFC3339, snapshot.Connections[i].Start)
			r, _ := time.Parse(time.RFC3339, snapshot.Connections[j].Start)
			return l.Before(r)
		})

		for _, c := range snapshot.Connections {
			host := c.Metadata.DstIP
			if c.Metadata.Host != "" {
				host = c.Metadata.Host
			}
			host = net.JoinHostPort(host, c.Metadata.DstPort)

			t, _ := time.Parse(time.RFC3339, c.Start)
			rows = append(rows, []interface{}{host, c.Metadata.NetWork, c.Metadata.Type, strings.Join(c.Chain, " --> "), c.Rule, time.Since(t).Round(time.Second).String()})
		}

		t.AppendRows(rows)
		t.Render()
	}
}
