package search

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/blevesearch/bleve"
)

const (
	USER_SEARCH         = "Users"
	ORGANIZATION_SEARCH = "Organizations"
	TICKET_SEARCH       = "Tickets"
)

type FileService interface {
	ReadFile(fileName string) ([]byte, error)
}

type Service struct {
	index bleve.Index
	fs    FileService
}

func New(fs FileService) *Service {
	svc := &Service{
		fs: fs,
	}
	return svc
}

func (svc *Service) unmarshallJson(searchType string) ([]map[string]interface{}, error) {
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

	jsonBytes, err := svc.fs.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	s := make([]map[string]interface{}, 0)

	err = json.Unmarshal(jsonBytes, &s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (svc *Service) Init(searchType string) error {
	s, err := svc.unmarshallJson(searchType)
	if err != nil {
		return err
	}

	mapping := bleve.NewIndexMapping()
	svc.index, err = bleve.NewMemOnly(mapping)
	if err != nil {
		return err
	}

	for _, m := range s {
		svc.index.Index(fmt.Sprintf("%s", m["_id"]), m)
	}
	return nil
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
			result = append(result, fmt.Sprintf("%s=%v", k, v))
		}
	} else {
		result = append(result, "No results found")
	}

	return strings.Join(result, "\n"), nil
}

func (svc *Service) ListFields() string {
	return ""
}
