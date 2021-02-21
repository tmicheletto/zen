package search

import (
	"encoding/json"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/google/uuid"
	"github.com/prometheus/common/log"
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

var USER_FIELDS = [...]string{"_id", "url", "email", "external_id", "tags", "created_at", "name", "organization_id", "organizations", "phone", "signature", "alias", "active", "verified", "shared", "locale", "timezone", "last_login_at", "suspended", "role"}
var ORGANIZATION_FIELDS = [...]string{"_id", "url", "external_id", "tags", "created_at", "name", "domain_names", "details", "shared_tickets"}
var TICKET_FIELDS = [...]string{"_id", "url", "external_id", "tags", "created_at", "type", "subject", "description", "priority", "status", "submitter_id", "assignee_id", "organization_id", "has_incidents", "due_at", "via"}

type FileService interface {
	ReadFile(fileName string) ([]byte, error)
}

type User struct {
	Id json.Number `json:"_id"`
	Name string `json:"name"`
	DocType DocType `json:"doc_type"`
	OrganizationId json.Number `json:"organization_id"`
	Organization Organization `json:"organization"`
	SubmittedTickets []Ticket `json:"submitted_tickets"`
}

type Organization struct {
	Id json.Number `json:"_id"`
	Name string `json:"name"`
	Users []User `json:"users"`
	DocType DocType `json:"doc_type"`
	Tickets []Ticket `json:"tickets"`
}

type Ticket struct {
	Id string `json:"_id"`
	DocType DocType `json:"doc_type"`
	Submitter User `json:"submitter"`
	Organization Organization `json:"organization"`
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

	users, err := svc.unmarshalUsers()
	if err != nil {
		return err
	}

	orgs, err := svc.unmarshalOrganizations()
	if err != nil {
		return err
	}

	tickets, err := svc.unmarshalTickets()
	if err != nil {
		return err
	}

	if err = svc.buildIndex(users, orgs, tickets); err != nil {
		return err
	}
	return nil
}

func(svc *Service) buildIndex(users []User, organizations []Organization, tickets []Ticket) error {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.TypeField = "doc_type"
	indexMapping.DefaultAnalyzer = "en"

	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "en"

	userMapping := bleve.NewDocumentMapping()
	for _, f := range USER_FIELDS {
		userMapping.AddFieldMappingsAt(f, textFieldMapping)
	}

	orgMapping := bleve.NewDocumentMapping()
	for _, f := range ORGANIZATION_FIELDS {
		orgMapping.AddFieldMappingsAt(f, textFieldMapping)
	}

	ticketMapping := bleve.NewDocumentMapping()
	for _, f := range TICKET_FIELDS {
		ticketMapping.AddFieldMappingsAt(f, textFieldMapping)
	}

	userMapping.AddSubDocumentMapping("organization", orgMapping)

	indexMapping.AddDocumentMapping("user", userMapping)
	indexMapping.AddDocumentMapping("organization", orgMapping)
	indexMapping.AddDocumentMapping("ticket", ticketMapping)

	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return err
	}

	for _, user := range users {
		user = buildUserGraph(user, organizations, tickets)
		log.Info(user.Organization)
		user.DocType = "user"

		index.Index(uuid.NewString(), user)
	}

	for _, org := range organizations {
		org.DocType = "organization"
		index.Index(uuid.NewString(), org)
	}

	for _, ticket := range tickets {
		ticket.DocType = "ticket"
		index.Index(uuid.NewString(), ticket)
	}
	svc.index = index
	return nil
}

func (svc *Service) Search(searchTerm string, searchValue string) ([]map[string]interface{}, error) {
	query := bleve.NewMatchQuery(searchValue)
	query.SetField(searchTerm)
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"*"}
	searchRequest.IncludeLocations = true
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

func (svc *Service) readFile(fileName string) ([]byte, error) {
	jsonBytes, err := svc.fs.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

func (svc *Service) unmarshalUsers() ([]User, error) {
	fileName := "./data/users.json"

	b, err := svc.readFile(fileName)
	if err != nil {
		return nil, err
	}

	var users []User
	if err = json.Unmarshal(b, &users); err != nil {
		return nil, err
	}

	for _, user := range users {
		user.DocType = USER_DOC_TYPE
	}
	return users, nil
}

func (svc *Service) unmarshalOrganizations() ([]Organization, error) {
	fileName := "./data/organizations.json"

	b, err := svc.readFile(fileName)
	if err != nil {
		return nil, err
	}

	var orgs []Organization
	if err = json.Unmarshal(b, &orgs); err != nil {
		return nil, err
	}

	for _, org := range orgs {
		org.DocType = ORGANIZATION_DOC_TYPE
	}
	return orgs, nil
}

func (svc *Service) unmarshalTickets() ([]Ticket, error) {
	fileName := "./data/tickets.json"

	b, err := svc.readFile(fileName)
	if err != nil {
		return nil, err
	}

	var tickets []Ticket
	if err = json.Unmarshal(b, &tickets); err != nil {
		return nil, err
	}

	for _, ticket := range tickets {
		ticket.DocType =  TICKET_DOC_TYPE
	}
	return tickets, nil
}

func buildUserGraph(user User, organizations []Organization, tickets []Ticket) User {
	for _, org := range organizations {
		if user.OrganizationId == org.Id {
			user.Organization = org
		}
	}
	return user
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
