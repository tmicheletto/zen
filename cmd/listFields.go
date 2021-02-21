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
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tmicheletto/zen/internal/file"
	"github.com/tmicheletto/zen/internal/search"
	"log"
)

// listFieldsCmd represents the listFields command
var listFieldsCmd = &cobra.Command{
	Use:   "list-fields",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
