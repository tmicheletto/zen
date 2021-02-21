package cmd

import (
	"fmt"
	"log"

	"github.com/tmicheletto/zen/internal/file"

	"github.com/manifoldco/promptui"
	"github.com/tmicheletto/zen/internal/search"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Searches the Zendesk database",
	Run: func(cmd *cobra.Command, args []string) {
		fs := file.New()
		svc := search.New(fs)

		searchTypePrompt := promptui.Select{
			Label: "What would you like to search for?",
			Items: []string{string(search.USER_SEARCH), string(search.TICKET_SEARCH), string(search.ORGANIZATION_SEARCH)},
		}
		_, searchType, err := searchTypePrompt.Run()
		if err != nil {
			log.Fatal("Prompt failed %v\n", err)
			return
		}

		err = svc.Init(search.Type(searchType))
		if err != nil {
			log.Fatal(err)
			return
		}

		searchTermPrompt := promptui.Select{
			Label: "Search term",
			Items: svc.ListFields(),
		}

		_, searchTerm, err := searchTermPrompt.Run()
		if err != nil {
			log.Fatal(err)
			return
		}

		searchValuePrompt := promptui.Prompt{
			Label: "Search value",
		}
		searchValue, err := searchValuePrompt.Run()
		if err != nil {
			log.Fatal(err)
			return
		}

		results, err := svc.Search(searchTerm, searchValue)
		if err != nil {
			log.Fatal(err)
			return
		}

		l := list.NewWriter()

		if len(results) > 0 {
			for i := 0; i < 10 && i < len(results); i++ {
				result := results[i]
				l.AppendItem(fmt.Sprintf("Result %d", i))
				l.Indent()
				for k, v := range result {
					l.AppendItem(fmt.Sprintf("%s: %v", k, v))
				}
				l.UnIndent()
			}

		} else {
			l.AppendItem("No results found")
		}
		fmt.Println(l.Render())
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
