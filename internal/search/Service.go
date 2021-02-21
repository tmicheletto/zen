package search

import (
	"encoding/json"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/mapping"
	"github.com/google/uuid"
	"strconv"
)

type Type string

const (
	USER_SEARCH         Type = "Users"
	ORGANIZATION_SEARCH Type = "Organizations"
	TICKET_SEARCH       Type = "Tickets"
)

type DocType string

const (
	USER_DOC_TYPE         DocType = "user"
	ORGANIZATION_DOC_TYPE DocType = "organization"
	TICKET_DOC_TYPE       DocType = "ticket"
)

var USER_FIELDS = [...]string{"_id", "url", "email", "external_id", "tags", "created_at", "name", "organization_id", "phone", "signature", "alias", "active", "verified", "shared", "locale", "timezone", "last_login_at", "suspended", "role"}
var ORGANIZATION_FIELDS = [...]string{"_id", "url", "external_id", "tags", "created_at", "name", "domain_names", "details", "shared_tickets"}
var TICKET_FIELDS = [...]string{"_id", "url", "external_id", "tags", "created_at", "type", "subject", "description", "priority", "status", "submitter_id", "assignee_id", "organization_id", "has_incidents", "due_at", "via"}

type FileService interface {
	ReadFile(fileName string) ([]byte, error)
}

type Service struct {
	index      bleve.Index
	searchType Type
	fs         FileService
}

func New(fs FileService) *Service {
	svc := &Service{
		fs: fs,
	}
	return svc
}

func (svc *Service) Init(searchType Type) error {
	svc.searchType = searchType

	indexMapping := bleve.NewIndexMapping()
	indexMapping.TypeField = "doc_type"
	indexMapping.DefaultAnalyzer = "en"

	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return err
	}

	index, err = svc.buildUserIndex(index, indexMapping)
	if err != nil {
		return err
	}

	index, err = svc.buildOrganizationIndex(index, indexMapping)
	if err != nil {
		return err
	}

	index, err = svc.buildTicketIndex(index, indexMapping)
	if err != nil {
		return err
	}
	svc.index = index
	return nil
}

func (svc *Service) Search(searchTerm string, searchValue string) ([]map[string]interface{}, error) {
	query := bleve.NewMatchQuery(searchValue)
	query.SetField(searchTerm)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"*"}
	searchResult, err := svc.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	for _, result := range searchResult.Hits {
			//if id := result.Fields["organization_id"]; id != nil {
			//	org, err := svc.getOrganisation(id.(string))
			//	if err != nil {
			//		return nil, err
			//	}
			//	if org != nil {
			//		result.Fields["organization_name"] = org["name"]
			//	}
			//}
			//tickets, err := svc.getTickets(result.Fields["_id"].(string))
			//if err != nil {
			//	return nil, err
			//}
			//for i, t := range tickets {
			//	result.Fields[fmt.Sprintf("ticket_%v", i)] = t["subject"]
			//}
		if searchTypeToDocType(svc.searchType) ==  DocType(result.Fields["doc_type"].(string)) {
			results = append(results, result.Fields)
		}
	}
	return results, nil
}

func (svc *Service) Fields() []string {
	var fields []string
	switch svc.searchType {
	case USER_SEARCH:
		fields = USER_FIELDS[:]
		break
	case ORGANIZATION_SEARCH:
		fields = ORGANIZATION_FIELDS[:]
		break
	case TICKET_SEARCH:
		fields = TICKET_FIELDS[:]
		break
	}
	return fields
}

func (svc *Service) unmarshallJson(fileName string) ([]map[string]interface{}, error) {
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

func (svc *Service) buildUserIndex(index bleve.Index, indexMapping *mapping.IndexMappingImpl) (bleve.Index, error) {
	fileName := "./data/users.json"

	items, err := svc.unmarshallJson(fileName)
	if err != nil {
		return nil, err
	}

	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "en"

	userMapping := bleve.NewDocumentMapping()

	for _, f := range USER_FIELDS {
		userMapping.AddFieldMappingsAt(f, textFieldMapping)
	}

	indexMapping.AddDocumentMapping("user", userMapping)

	intFieldsToConvert := []string{"_id", "organization_id"}
	boolFieldsToConvert := []string{"active", "verified", "shared", "suspended"}

	for _, item := range items {
		item = convertIntFieldsToText(item, intFieldsToConvert)
		item = convertBoolFieldsToText(item, boolFieldsToConvert)

		item["doc_type"] = "user"
		index.Index(uuid.NewString(), item)
	}
	return index, nil
}

func (svc *Service) buildOrganizationIndex(index bleve.Index, indexMapping *mapping.IndexMappingImpl) (bleve.Index, error) {
	fileName := "./data/organizations.json"

	items, err := svc.unmarshallJson(fileName)
	if err != nil {
		return nil, err
	}

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "en"

	orgMapping := bleve.NewDocumentMapping()

	for _, f := range ORGANIZATION_FIELDS {
		orgMapping.AddFieldMappingsAt(f, textFieldMapping)
	}

	indexMapping.AddDocumentMapping("organization", orgMapping)

	intFieldsToConvert := []string{"_id"}
	boolFieldsToConvert := []string{"shared_tickets"}

	for _, item := range items {
		item = convertIntFieldsToText(item, intFieldsToConvert)
		item = convertBoolFieldsToText(item, boolFieldsToConvert)

		item["doc_type"] = "organization"
		index.Index(uuid.NewString(), item)
	}
	return index, nil
}

func (svc *Service) buildTicketIndex(index bleve.Index, indexMapping *mapping.IndexMappingImpl) (bleve.Index, error) {
	fileName := "./data/tickets.json"

	items, err := svc.unmarshallJson(fileName)
	if err != nil {
		return nil, err
	}

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = keyword.Name

	ticketMapping := bleve.NewDocumentMapping()

	for _, f := range TICKET_FIELDS {
		ticketMapping.AddFieldMappingsAt(f, textFieldMapping)
	}

	indexMapping.AddDocumentMapping("ticket", ticketMapping)

	intFieldsToConvert := []string{"organization_id", "submitter_id", "assignee_id"}
	boolFieldsToConvert := []string{"has_incidents"}

	for _, item := range items {
		item["_id"] = item["_id"].(string)
		item = convertIntFieldsToText(item, intFieldsToConvert)
		item = convertBoolFieldsToText(item, boolFieldsToConvert)

		item["doc_type"] = "ticket"
		index.Index(uuid.NewString(), item)
	}
	return index, nil
}

func convertBoolFieldsToText(m map[string]interface{}, fieldNamesToConvert []string) map[string]interface{} {
	for _, f := range fieldNamesToConvert {
		if v := m[f]; v != nil {
			m[f] = strconv.FormatBool(v.(bool))
		}
	}
	return m
}

func convertIntFieldsToText(m map[string]interface{}, fieldNamesToConvert []string) map[string]interface{} {
	for _, f := range fieldNamesToConvert {
		if v := m[f]; v != nil {
			m[f] = strconv.Itoa(int(v.(float64)))
		}
	}
	return m
}

func (svc *Service) getOrganisation(id string) (map[string]interface{}, error) {

	var result map[string]interface{}
	query := bleve.NewTermQuery(id)
	query.SetField("_id")
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"*"}
	searchResult, err := svc.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(searchResult.Hits) > 0 {
		result = searchResult.Hits[0].Fields
	}
	return result, nil
}

func (svc *Service) getTickets(userId string) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)
	query := bleve.NewTermQuery(userId)
	query.SetField("assignee_id")
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"*"}
	searchResult, err := svc.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	for _, h := range searchResult.Hits {
		result = append(result, h.Fields)
	}
	return result, nil
}

func searchTypeToDocType(searchType Type) DocType {
	var docType DocType
	switch searchType {
	case USER_SEARCH:
		docType = USER_DOC_TYPE
		break
	case ORGANIZATION_SEARCH:
		docType = ORGANIZATION_DOC_TYPE
		break
	case TICKET_SEARCH:
		docType = TICKET_DOC_TYPE
		break
	}
	return docType
}
