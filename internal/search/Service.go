package search

import (
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/google/uuid"
	"reflect"
)

type Type string

const (
	USER_SEARCH         Type = "Users"
	ORGANIZATION_SEARCH Type = "Organizations"
	TICKET_SEARCH       Type = "Tickets"
)

const DOC_TYPE_FIELD_NAME = "DocType"

type DocType string

const (
	USER_DOC_TYPE         DocType = "user"
	ORGANIZATION_DOC_TYPE DocType = "organization"
	TICKET_DOC_TYPE       DocType = "ticket"
)

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

type User struct {
	Id               json.Number `json:"_id"`
	Url              string      `json:"url"`
	ExternalId       string      `json:"external_id"`
	Name             string      `json:"name"`
	Alias            string      `json:"alias"`
	CreatedAt        string      `json:"created_at"`
	Active           bool        `json:"active"`
	Shared           bool        `json:"shared"`
	Verified         bool        `json:"verified"`
	Locale           string      `json:"locale"`
	TimeZone         string      `json:"timezone"`
	LastLoginAt      string      `json:"last_login_at"`
	Email            string      `json:"email"`
	Phone            string      `json:"phone"`
	Signature        string      `json:"signature"`
	OrganizationId   json.Number `json:"organization_id"`
	Tags             []string    `json:"tags"`
	Suspended        bool        `json:"suspended"`
	Role             string      `json:"role"`
	DocType          DocType
	Organization     Organization
	SubmittedTickets []Ticket
	AssignedTickets  []Ticket
}

func getFields(target interface{}) []string {
	val := reflect.ValueOf(target).Elem()

	fields := make([]string, 0)
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		field := tag.Get("json")
		if field != "" && field != DOC_TYPE_FIELD_NAME {
			fields = append(fields, field)
		}
	}
	return fields
}

func buildUserMapping(keywordMapping *mapping.FieldMapping, textFieldMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	userMapping := bleve.NewDocumentMapping()
	userMapping.AddFieldMappingsAt("_id", keywordMapping)
	userMapping.AddFieldMappingsAt("url", textFieldMapping)
	userMapping.AddFieldMappingsAt("external_id", textFieldMapping)
	userMapping.AddFieldMappingsAt("name", textFieldMapping)
	userMapping.AddFieldMappingsAt("alias", textFieldMapping)
	userMapping.AddFieldMappingsAt("created_at", textFieldMapping)
	userMapping.AddFieldMappingsAt("active", textFieldMapping)
	userMapping.AddFieldMappingsAt("verified", keywordMapping)
	userMapping.AddFieldMappingsAt("shared", keywordMapping)
	userMapping.AddFieldMappingsAt("locale", textFieldMapping)
	userMapping.AddFieldMappingsAt("timezone", textFieldMapping)
	userMapping.AddFieldMappingsAt("last_login_at", textFieldMapping)
	userMapping.AddFieldMappingsAt("email", textFieldMapping)
	userMapping.AddFieldMappingsAt("phone", textFieldMapping)
	userMapping.AddFieldMappingsAt("signature", textFieldMapping)
	userMapping.AddFieldMappingsAt("organization_id", keywordMapping)
	userMapping.AddFieldMappingsAt("tags", textFieldMapping)
	userMapping.AddFieldMappingsAt("suspended", keywordMapping)
	userMapping.AddFieldMappingsAt("role", textFieldMapping)

	return userMapping
}

type Organization struct {
	Id            json.Number `json:"_id"`
	Url           string      `json:"url"`
	ExternalId    string      `json:"external_id"`
	Name          string      `json:"name"`
	DomainNames   []string    `json:"domain_names"`
	CreatedAt     string      `json:"created_at"`
	Details       string      `json:"details"`
	SharedTickets bool        `json:"shared_tickets"`
	Tags          []string    `json:"tags"`
	Users         []User
	DocType       DocType
	Tickets       []Ticket
}

func buildOrganizationMapping(keywordMapping *mapping.FieldMapping, textFieldMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	orgMapping := bleve.NewDocumentMapping()
	orgMapping.AddFieldMappingsAt("_id", keywordMapping)
	orgMapping.AddFieldMappingsAt("name", textFieldMapping)
	orgMapping.AddFieldMappingsAt("external_id", keywordMapping)
	orgMapping.AddFieldMappingsAt("domain_names", textFieldMapping)
	orgMapping.AddFieldMappingsAt("created_at", textFieldMapping)
	orgMapping.AddFieldMappingsAt("details", textFieldMapping)
	orgMapping.AddFieldMappingsAt("shared_tickets", textFieldMapping)
	orgMapping.AddFieldMappingsAt("tags", textFieldMapping)

	return orgMapping
}

type Ticket struct {
	Id             string   `json:"_id"`
	Url            string   `json:"url"`
	ExternalId     string   `json:"external_id"`
	CreatedAt      string   `json:"created_at"`
	Type           string   `json:"type"`
	Subject        string   `json:"subject"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`
	Status         string   `json:"status"`
	Tags           []string `json:"tags"`
	HasIncidents   bool     `json:"has_incidents"`
	DueAt          string   `json:"due_at"`
	Via            string   `json:"via"`
	DocType        DocType
	Submitter      User
	SubmitterId    json.Number `json:"submitter_id"`
	AssigneeId     json.Number `json:"assignee_id"`
	Assignee       User
	OrganizationId json.Number `json:"organization_id"`
	Organization   Organization
}

func buildTicketMapping(keywordMapping *mapping.FieldMapping, textFieldMapping *mapping.FieldMapping) *mapping.DocumentMapping {
	ticketMapping := bleve.NewDocumentMapping()
	ticketMapping.AddFieldMappingsAt("_id", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("url", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("external_id", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("created_at", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("type", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("subject", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("description", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("priority", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("status", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("tags", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("has_incidents", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("due_at", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("via", textFieldMapping)
	ticketMapping.AddFieldMappingsAt("submitter_id", keywordMapping)
	ticketMapping.AddFieldMappingsAt("assignee_id", keywordMapping)
	ticketMapping.AddFieldMappingsAt("organization_id", keywordMapping)

	return ticketMapping
}

func (svc *Service) buildIndex(users []User, organizations []Organization, tickets []Ticket) error {
	indexMapping := bleve.NewIndexMapping()
	indexMapping.TypeField = "DocType"
	indexMapping.DefaultAnalyzer = "en"

	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = keyword.Name

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = en.AnalyzerName

	userMapping := buildUserMapping(keywordMapping, textFieldMapping)
	orgMapping := buildOrganizationMapping(keywordMapping, textFieldMapping)
	ticketMapping := buildTicketMapping(keywordMapping, textFieldMapping)

	indexMapping.AddDocumentMapping(string(USER_DOC_TYPE), userMapping)
	indexMapping.AddDocumentMapping(string(ORGANIZATION_DOC_TYPE), orgMapping)
	indexMapping.AddDocumentMapping(string(TICKET_DOC_TYPE), ticketMapping)

	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return err
	}

	for _, user := range users {
		user.DocType = USER_DOC_TYPE
		user = buildUserGraph(user, organizations, tickets)
		index.Index(uuid.NewString(), user)
	}

	for _, org := range organizations {
		org.DocType = ORGANIZATION_DOC_TYPE
		index.Index(uuid.NewString(), org)
	}

	for _, ticket := range tickets {
		ticket.DocType = TICKET_DOC_TYPE
		ticket = buildTicketGraph(ticket, organizations, users)
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
	searchResult, err := svc.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	for _, result := range searchResult.Hits {
		if searchTypeToDocType(svc.searchType) == DocType(result.Fields[DOC_TYPE_FIELD_NAME].(string)) {
			displayFields := svc.ListFields()
			m := make(map[string]interface{})
			for k, v := range result.Fields {
				if contains(displayFields, k) {
					m[k] = v
				}
			}

			switch svc.searchType {
			case USER_SEARCH:
				m = mapAdditionalUserFields(result.Fields, m)
				break
			case TICKET_SEARCH:
				m = mapAdditionalTicketFields(result.Fields, m)
				break
			}
			results = append(results, m)
		}
	}

	return results, nil
}

func mapAdditionalUserFields(searchResults map[string]interface{}, result map[string]interface{}) map[string]interface{} {
	result["organization"] = searchResults["Organization.name"]
	assignedTickets := searchResults["AssignedTickets.subject"].([]interface{})
	for i, t := range assignedTickets {
		result[fmt.Sprintf("assigned_ticket_%d", i)] = t
	}

	submittedTickets := searchResults["SubmittedTickets.subject"].([]interface{})
	for i, t := range submittedTickets {
		result[fmt.Sprintf("submitted_ticket_%d", i)] = t
	}
	return result
}

func mapAdditionalTicketFields(searchResults map[string]interface{}, result map[string]interface{}) map[string]interface{} {
	result["organization"] = searchResults["Organization.name"]
	result["assignee"] = searchResults["Assignee.name"]
	result["submitter"] = searchResults["Submitter.name"]
	return result
}

func (svc *Service) ListFields() []string {
	var fields []string
	switch svc.searchType {
	case USER_SEARCH:
		user := &User{}
		fields = getFields(user)
		break
	case ORGANIZATION_SEARCH:
		org := &Organization{}
		fields = getFields(org)
		break
	case TICKET_SEARCH:
		ticket := &Ticket{}
		fields = getFields(ticket)
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
	return tickets, nil
}

func buildUserGraph(user User, organizations []Organization, tickets []Ticket) User {
	for _, org := range organizations {
		if user.OrganizationId == org.Id {
			user.Organization = org
		}
	}

	submittedTickets := make([]Ticket, 0)
	assignedTickets := make([]Ticket, 0)
	for _, ticket := range tickets {
		if user.Id == ticket.SubmitterId {
			submittedTickets = append(submittedTickets, ticket)
		}
		if user.Id == ticket.AssigneeId {
			assignedTickets = append(assignedTickets, ticket)
		}
	}
	user.SubmittedTickets = submittedTickets
	user.AssignedTickets = assignedTickets
	return user
}

func buildTicketGraph(ticket Ticket, organizations []Organization, users []User) Ticket {
	for _, org := range organizations {
		if ticket.OrganizationId == org.Id {
			ticket.Organization = org
		}
	}
	for _, user := range users {
		if ticket.AssigneeId == user.Id {
			ticket.Assignee = user
		}
		if ticket.SubmitterId == user.Id {
			ticket.Submitter = user
		}
	}
	return ticket
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

func contains(slice []string, element string) bool {
	for _, e := range slice {
		if e == element {
			return true
		}
	}
	return false
}
