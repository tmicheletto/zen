package cmd

import (
	"fmt"
	"log"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tmicheletto/zen/internal/file"
	"github.com/tmicheletto/zen/internal/search"
)

// listFieldsCmd represents the listFields command
var listFieldsCmd = &cobra.Command{
	Use:   "list-fields",
	Short: "Lists the available fields to search",
	Run: func(cmd *cobra.Command, args []string) {
		searchTypePrompt := promptui.Select{
			Label: "What would you like to search for?",
			Items: []string{string(search.USER_SEARCH), string(search.TICKET_SEARCH), string(search.ORGANIZATION_SEARCH)},
		}
		_, searchType, err := searchTypePrompt.Run()
		if err != nil {
			log.Fatal("Prompt failed %v\n", err)
			return
		}

		fs := file.New()
		svc := search.New(fs)
		err = svc.Init(search.Type(searchType))
		if err != nil {
			log.Fatal(err)
			return
		}

		l := list.NewWriter()
		for _, f := range svc.ListFields() {
			l.AppendItem(f)
		}
		fmt.Println(l.Render())
	},
}

func init() {
	rootCmd.AddCommand(listFieldsCmd)
}
