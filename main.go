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
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	args := strings.Split(in.TextBeforeCursor(), " ")
	w := in.GetWordBeforeCursor()

	// first class command
	if len(args) <= 1 {
		return prompt.FilterHasPrefix(
			[]prompt.Suggest{
				{Text: "server", Description: "manage remote clash server"},
				{Text: "now", Description: "show selected clash server"},
				{Text: "use", Description: "change selected clash server"},
				{Text: "ping", Description: "check clash servers alive"},
				{Text: "traffic", Description: "get clash traffic"},
				{Text: "connections", Description: "get clash all connections"},
			},
			w,
			true,
		)
	}

	switch args[0] {
	case "server":
		return prompt.FilterHasPrefix(
			[]prompt.Suggest{
				{Text: "ls", Description: "list all server"},
				{Text: "add", Description: "add new server"},
			},
			args[1],
			true,
		)
	case "use":
		cfg, err := common.ReadCfg()
		if err != nil {
			return []prompt.Suggest{}
		}

		suggests := []prompt.Suggest{}
		for key := range cfg.Servers {
			suggests = append(suggests, prompt.Suggest{Text: key})
		}

		return suggests
	}

	return []prompt.Suggest{}
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
