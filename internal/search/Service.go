package search

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/blevesearch/bleve"
)

const (
	USER_SEARCH = "Users"
	ORGANIZATION_SEARCH = "Organizations"
	TICKET_SEARCH = "Tickets"
)

type Service struct {
	index bleve.Index
}

func NewSearchService(searchType string) (*Service, error) {
	index, err := openIndex(searchType)
	if err != nil {
		return nil, err
	}
	svc := &Service{
		index: index,
	}
	return svc, nil
}

func unmarshallJson(searchType string) ([]map[string]interface{}, error) {
	var fileName string

	switch searchType {
	case USER_SEARCH:
		fileName = "./data/users.json"
		break
	case ORGANIZATION_SEARCH:
		fileName = "./data/organizations.json"
		break
	case TICKET_SEARCH:
		fileName = "./data/tickets.json"
		break
	default:
		return nil, errors.Errorf("Unsupported search type: %s", searchType)
	}


	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	s := make([]map[string]interface{}, 0)

	err = json.Unmarshal(byteValue, &s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func openIndex(searchType string) (bleve.Index, error) {
	var indexFile string

	switch searchType {
	case USER_SEARCH:
		indexFile = "./data/idx/users.idx"
		break
	case ORGANIZATION_SEARCH:
		indexFile = "./data/idx/organizations.idx"
		break
	case TICKET_SEARCH:
		indexFile = "./data/idx/tickets.idx"
		break
	default:
		return nil, errors.Errorf("Unsupported search type: %s", searchType)
	}

	index, err := bleve.Open(indexFile)
	if err == nil {
		return index, nil
	}

	s, err := unmarshallJson(searchType)
	if err != nil {
		return nil, err
	}

	mapping := bleve.NewIndexMapping()
	index, err = bleve.New(indexFile, mapping)
	if err != nil {
		return nil, err
	}

	for _, m := range s {
		index.Index(fmt.Sprintf("%s", m["_id"]), m)
	}
	return index, nil
}

func (svc *Service) Search(searchTerm string, searchValue string) (string, error) {
	query := bleve.NewMatchQuery(fmt.Sprintf("%s=%s", searchTerm, searchValue))
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"*"}
	searchResult, err := svc.index.Search(searchRequest)
	if err != nil {
		return "", err
	}

	result := make([]string, 0)
	if len(searchResult.Hits) > 0 {
		for k, v := range searchResult.Hits[0].Fields {
			result = append(result, fmt.Sprintf("%s		%v", k, v))
		}
	} else {
		result = append(result, "No results found")
	}

	return strings.Join(result, "\n"), nil
}

func (svc *Service) ListFields() string {
	return ""
}
