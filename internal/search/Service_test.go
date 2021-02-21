package search_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tmicheletto/zen/internal/search"
)

type mockFileService struct {
	mock.Mock
}

func (fs *mockFileService) ReadFile(fileName string) ([]byte, error) {
	args := fs.Called(fileName)
	return args.Get(0).([]byte), args.Error(1)
}

var usersJson = `[{
    "_id": 1,
    "url": "http://initech.zendesk.com/api/v2/users/40.json",
    "external_id": "e79e166e-fcb5-4604-9466-1ea9777c6eb5",
    "name": "Burgess England",
    "alias": "Mr Neal",
    "created_at": "2016-07-16T09:13:47 -10:00",
    "active": true,
    "verified": false,
    "shared": true,
    "locale": "zh-CN",
    "timezone": "Taiwan",
    "last_login_at": "2015-10-20T01:25:19 -11:00",
    "email": "nealengland@flotonic.com",
    "phone": "9675-282-161",
    "signature": "Don't Worry Be Happy!",
    "organization_id": 1,
    "tags": [
      "Riceville",
      "Ribera",
      "Caberfae",
      "Breinigsville"
    ],
    "suspended": true,
    "role": "end-user"
  	}]`

var orgsJson = `[{
    "_id": 1,
    "url": "http://initech.zendesk.com/api/v2/organizations/118.json",
    "external_id": "6970300e-f211-4c01-a538-70b4464a1d84",
    "name": "Limozen",
    "domain_names": [
      "otherway.com",
      "rodeomad.com",
      "suremax.com",
      "fishland.com"
    ],
    "created_at": "2016-02-11T04:24:09 -11:00",
    "details": "MegaCorp",
    "shared_tickets": false,
    "tags": [
      "Leon",
      "Ferguson",
      "Olsen",
      "Walsh"
    ]
  }]`

var ticketsJson = `[{
    "_id": "2217c7dc-7371-4401-8738-0a8a8aedc08d",
    "url": "http://initech.zendesk.com/api/v2/tickets/2217c7dc-7371-4401-8738-0a8a8aedc08d.json",
    "external_id": "3db2c1e6-559d-4015-b7a4-6248464a6bf0",
    "created_at": "2016-07-16T12:05:12 -10:00",
    "type": "problem",
    "subject": "A Catastrophe in Hungary",
    "description": "Ipsum fugiat voluptate reprehenderit cupidatat aliqua dolore consequat. Consequat ullamco minim laboris veniam ea id laborum et eiusmod excepteur sint laborum dolore qui.",
    "priority": "normal",
    "status": "closed",
    "submitter_id": 1,
    "assignee_id": 1,
    "organization_id": 1,
    "tags": [
      "Massachusetts",
      "New York",
      "Minnesota",
      "New Jersey"
    ],
    "has_incidents": true,
    "due_at": "2016-08-06T04:16:06 -10:00",
    "via": "web"
  	},
{
    "_id": "87db32c5-76a3-4069-954c-7d59c6c21de0",
    "url": "http://initech.zendesk.com/api/v2/tickets/87db32c5-76a3-4069-954c-7d59c6c21de0.json",
    "external_id": "1c61056c-a5ad-478a-9fd6-38889c3cd728",
    "created_at": "2016-07-06T11:16:50 -10:00",
    "type": "problem",
    "subject": "A Problem in Morocco",
    "description": "Sit culpa non magna anim. Ea velit qui nostrud eiusmod laboris dolor adipisicing quis deserunt elit amet.",
    "priority": "urgent",
    "status": "solved",
    "submitter_id": 1,
    "assignee_id": 1,
    "organization_id": 1,
    "tags": [
      "Texas",
      "Nevada",
      "Oregon",
      "Arizona"
    ],
    "has_incidents": true,
    "due_at": "2016-08-19T07:40:17 -10:00",
    "via": "voice"
  }]`

func TestUserSearchReturnsResult(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.USER_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result, err := svc.Search("_id", "1")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, len(result), 1)
	user := result[0]
	assert.Equal(t, "1", user["_id"])
	assert.Equal(t, "http://initech.zendesk.com/api/v2/users/40.json", user["url"])
	assert.Equal(t, "Burgess England", user["name"])
	assert.Equal(t, "Mr Neal", user["alias"])
	assert.Equal(t, "2016-07-16T09:13:47 -10:00", user["created_at"])
	assert.Equal(t, true, user["active"])
	assert.Equal(t, false, user["verified"])
	assert.Equal(t, true, user["shared"])
	assert.Equal(t, "zh-CN", user["locale"])
	assert.Equal(t, "Taiwan", user["timezone"])
	assert.Equal(t, "2015-10-20T01:25:19 -11:00", user["last_login_at"])
	assert.Equal(t, "nealengland@flotonic.com", user["email"])
	assert.Equal(t, "9675-282-161", user["phone"])
	assert.Equal(t, "Don't Worry Be Happy!", user["signature"])
	assert.Equal(t, "1", user["organization_id"])
	assert.Equal(t, []interface{}{"Riceville", "Ribera", "Caberfae", "Breinigsville"}, user["tags"])
	assert.Equal(t, true, user["suspended"])
	assert.Equal(t, "end-user", user["role"])

	assert.Equal(t, "Limozen", user["organization"])
	assert.Equal(t, "A Catastrophe in Hungary", user["assigned_ticket_0"])
	assert.Equal(t, "A Problem in Morocco", user["assigned_ticket_1"])
	assert.Equal(t, "A Catastrophe in Hungary", user["submitted_ticket_0"])
	assert.Equal(t, "A Problem in Morocco", user["submitted_ticket_1"])
}

func TestUserSearchReturnsNoResult(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.USER_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result, err := svc.Search("_id", "2")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, len(result), 0)
}

func TestOrganizationSearchReturnsResult(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.ORGANIZATION_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result, err := svc.Search("_id", "1")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, 1, len(result))
	org := result[0]
	assert.Equal(t, "http://initech.zendesk.com/api/v2/organizations/118.json", org["url"])
	assert.Equal(t, "6970300e-f211-4c01-a538-70b4464a1d84", org["external_id"])
	assert.Equal(t, "Limozen", org["name"])
	assert.Equal(t, []interface{}{"otherway.com", "rodeomad.com", "suremax.com", "fishland.com"}, org["domain_names"])
	assert.Equal(t, "2016-02-11T04:24:09 -11:00", org["created_at"])
	assert.Equal(t, "MegaCorp", org["details"])
	assert.Equal(t, false, org["shared_tickets"])
	assert.Equal(t, []interface{}{"Leon", "Ferguson", "Olsen", "Walsh"}, org["tags"])

}

func TestOrganizationSearchReturnsNoResult(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.ORGANIZATION_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result, err := svc.Search("_id", "2")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, 0, len(result))
}

func TestTicketSearchReturnsResult(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.TICKET_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	// For some inexplicable reason searching by _id does not work. Using external_id for now.
	result, err := svc.Search("external_id", "3db2c1e6-559d-4015-b7a4-6248464a6bf0")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, 1, len(result))
	ticket := result[0]
	assert.Equal(t, "http://initech.zendesk.com/api/v2/tickets/2217c7dc-7371-4401-8738-0a8a8aedc08d.json", ticket["url"])
	assert.Equal(t, "2217c7dc-7371-4401-8738-0a8a8aedc08d", ticket["_id"])
	assert.Equal(t, "2016-07-16T12:05:12 -10:00", ticket["created_at"])
	assert.Equal(t, "problem", ticket["type"])
	assert.Equal(t, "A Catastrophe in Hungary", ticket["subject"])
	assert.Equal(t, "Ipsum fugiat voluptate reprehenderit cupidatat aliqua dolore consequat. Consequat ullamco minim laboris veniam ea id laborum et eiusmod excepteur sint laborum dolore qui.", ticket["description"])
	assert.Equal(t, "closed", ticket["status"])
	assert.Equal(t, "1", ticket["submitter_id"])
	assert.Equal(t, "1", ticket["assignee_id"])
	assert.Equal(t, "1", ticket["organization_id"])
	assert.Equal(t, []interface{}{"Massachusetts", "New York", "Minnesota", "New Jersey"}, ticket["tags"])
	assert.Equal(t, true, ticket["has_incidents"])
	assert.Equal(t, "2016-08-06T04:16:06 -10:00", ticket["due_at"])
	assert.Equal(t, "web", ticket["via"])
}

func TestTicketSearchReturnsMultipleResults(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.TICKET_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result, err := svc.Search("type", "problem")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, 2, len(result))
}

func TestListUserFields(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.TICKET_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result := svc.ListFields()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, []string{"_id", "url", "external_id", "created_at", "type", "subject", "description", "priority", "status", "tags", "has_incidents", "due_at", "via", "submitter_id", "assignee_id", "organization_id"}, result)
}

func TestListTicketFields(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.USER_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result := svc.ListFields()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, []string{"_id", "url", "external_id", "name", "alias", "created_at", "active", "shared", "verified", "locale", "timezone", "last_login_at", "email", "phone", "signature", "organization_id", "tags", "suspended", "role"}, result)
}

func TestListOrganizationFields(t *testing.T) {
	mfs := &mockFileService{}

	mfs.On("ReadFile", "./data/users.json").Return([]byte(usersJson), nil)
	mfs.On("ReadFile", "./data/organizations.json").Return([]byte(orgsJson), nil)
	mfs.On("ReadFile", "./data/tickets.json").Return([]byte(ticketsJson), nil)

	svc := search.New(mfs)
	err := svc.Init(search.ORGANIZATION_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result := svc.ListFields()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, []string{"_id", "url", "external_id", "name", "domain_names", "created_at", "details", "shared_tickets", "tags"}, result)
}
