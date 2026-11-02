package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/neko233-com/image-cli/internal/agent"
)

func runAgent(args []string) error {
	fs := flag.NewFlagSet("agent", flag.ExitOnError)
	task := fs.String("task", "", "只打印一条配方 id")
	asJSON := fs.Bool("json", false, "输出规则和全部配方 JSON")
	_ = fs.Parse(args)
	if *task != "" {
		item, ok := agent.Find(*task)
		if !ok {
			return fmt.Errorf("unknown task %q", *task)
		}
		if *asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(item)
		}
		fmt.Printf("id: %s\nwhen: %s\ncommand:\n%s\n", item.ID, item.When, item.Command)
		if item.Notes != "" {
			fmt.Printf("notes: %s\n", item.Notes)
		}
		return nil
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(struct {
			Prompt string       `json:"prompt"`
			Tasks  []agent.Task `json:"tasks"`
		}{Prompt: agent.Prompt(), Tasks: agent.Tasks()})
	}
	fmt.Print(agent.Prompt())
	fmt.Println()
	fmt.Println("配方：")
	for _, item := range agent.Tasks() {
		fmt.Printf("- %s [%s] %s\n  %s\n", item.ID, item.Audience, item.When, item.Command)
		if item.Notes != "" {
			fmt.Printf("  %s\n", item.Notes)
		}
	}
	fmt.Println()
	fmt.Println("单条配方：image-cli agent -task game-ui ；全部 JSON：image-cli agent -json")
	return nil
}
