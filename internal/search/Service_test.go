package search_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tmicheletto/zen/internal/search"
	"testing"
)

type mockFileService struct {
	mock.Mock
}

func (fs *mockFileService) ReadFile(fileName string) ([]byte, error) {
	args := fs.Called(fileName)
	return args.Get(0).([]byte), args.Error(1)
}

func TestSomething(t *testing.T) {
	mfs := &mockFileService{}

	jsonString := `[{
    "_id": 40,
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
    "organization_id": 113,
    "tags": [
      "Riceville",
      "Ribera",
      "Caberfae",
      "Breinigsville"
    ],
    "suspended": true,
    "role": "end-user"
  	}]`

	mfs.On("ReadFile", "./data/users.json").Return([]byte(jsonString), nil)

	svc := search.New(mfs)
	err := svc.Init(search.USER_SEARCH)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	result, err := svc.Search("_id", "40")
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Contains(t, result, "_id=40")
	assert.Contains(t, result, "email=nealengland@flotonic.com")
}
