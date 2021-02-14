/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/tmicheletto/zen/internal/search"

	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		searchTypePrompt := promptui.Select{
			Label: "What would you like to search for?",
			Items: []string{search.USER_SEARCH, search.TICKET_SEARCH, search.ORGANIZATION_SEARCH},
		}
		_, searchType, err := searchTypePrompt.Run()

		if err != nil {
			log.Fatal("Prompt failed %v\n", err)
			return
		}

		svc, err := search.NewSearchService(searchType)
		if err != nil {
			log.Fatal(err)
			return
		}

		validate := func(input string) error {
			if len(input) == 0 {
				return errors.New("You must enter a value")
			}
			return nil
		}

		searchTermPrompt := promptui.Prompt{
			Label:    "Search term",
			Validate: validate,
		}

		searchTerm, err := searchTermPrompt.Run()
		if err != nil {
			log.Fatal(err)
			return
		}

		searchValuePrompt := promptui.Prompt{
			Label:    "Search value",
			Validate: validate,
		}
		searchValue, err := searchValuePrompt.Run()
		if err != nil {
			log.Fatal(err)
			return
		}

		result, err := svc.Search(searchTerm, searchValue)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
