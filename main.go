package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Dreamacro/clash-ctl/commands"
	"github.com/Dreamacro/clash-ctl/common"

	"github.com/c-bata/go-prompt"
	"github.com/jedib0t/go-pretty/v6/text"
)

var root = []common.Node{
	{Text: "now", Description: "show selected clash server"},
	{Text: "ping", Description: "check clash servers alive"},
	{Text: "traffic", Description: "get clash traffic"},
	{Text: "connections", Description: "get clash all connections"},
	{
		Text: "server", Description: "manage remote clash server",
		Children: []common.Node{
			{Text: "ls", Description: "list all server"},
			{Text: "add", Description: "add new server"},
		},
	},
	{
		Text: "proxy", Description: "manage remote clash proxy",
		Children: []common.Node{
			{
				Text: "ls", Description: "list all proxy",
				// Resolver: commands.ProxyListResolver,
			},
			{
				Text: "set", Description: "select a proxy from a group",
				Resolver: commands.ProxySetResolver,
			},
		},
	},
	{
		Text: "use", Description: "change selected clash server",
		Resolver: commands.UseServerResolover,
	},
}

func executor(in string) {
	in = strings.TrimSpace(in)

	blocks := strings.Split(in, " ")
	switch blocks[0] {
	case "exit":
		fmt.Println("Bye!")
		os.Exit(0)
	case "server":
		commands.HandleServerCommand(blocks[1:])
	case "now", "use", "ping":
		commands.HandleMiscCommand(blocks)
	case "traffic", "connections":
		commands.HandleCommonCommand(blocks)
	case "proxy":
		commands.HandleProxyCommand(blocks[1:])
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	args := strings.Split(in.TextBeforeCursor(), " ")
	n := root
	prefixIdx := 0

outside:
	for i := 0; i < len(args)-1; i++ {
		var next []common.Node
		keyword := args[prefixIdx]
		if keyword == "" {
			break
		}

		for _, nd := range n {
			if nd.Text == keyword {
				if nd.Resolver != nil {
					var step int
					step, n = nd.Resolver(args[prefixIdx+1:])
					prefixIdx += step
					break outside
				} else if len(nd.Children) != 0 {
					next = nd.Children
				}
				break
			}
		}

		prefixIdx++
		if next == nil {
			n = nil
			break
		}

		n = next
	}

	suggestions := []prompt.Suggest{}
	for _, sg := range n {
		suggestion := prompt.Suggest{Text: sg.Text, Description: sg.Description}
		suggestions = append(suggestions, suggestion)
	}

	return prompt.FilterHasPrefix(
		suggestions,
		args[prefixIdx],
		true,
	)
}

func main() {
	if err := common.Init(); err != nil {
		fmt.Println(text.FgRed.Sprint(err.Error()))
		return
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("clash-ctl"),
		prompt.OptionCompletionOnDown(),
		prompt.OptionShowCompletionAtStart(),
	)
	p.Run()
}
