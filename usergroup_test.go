package wavefront

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

type MockUserGroupClient struct {
	Client
	T *testing.T
}

type MockCrudUserGroupClient struct {
	Client
	T      *testing.T
	method string
}

func (m *MockUserGroupClient) Do(req *http.Request) (io.ReadCloser, error) {
	body, _ := ioutil.ReadAll(req.Body)
	search := SearchParams{}
	err := json.Unmarshal(body, &search)
	if err != nil {
		m.T.Fatal(err)
	}

	response, err := ioutil.ReadFile("./fixtures/search-usergroup-response.json")
	if err != nil {
		m.T.Fatal(err)
	}
	return ioutil.NopCloser(bytes.NewReader(response)), nil
}

func TestUserGroups_Find(t *testing.T) {
	baseurl, _ := url.Parse("http://testing.wavefront.com")
	g := &UserGroups{
		client: &MockUserGroupClient{
			Client: Client{
				Config:     &Config{Token: "1234-5678-9977"},
				BaseURL:    baseurl,
				httpClient: http.DefaultClient,
				debug:      true,
			},
			T: t,
		},
	}

	userGroups, err := g.Find(nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(userGroups) != 1 {
		t.Errorf("expected one UserGroup returned, got %d", len(userGroups))
	}

	if *userGroups[0].ID != "12345678-1234-5678-9977-123456789111" {
		t.Errorf("expected first UserGroup to id to be 12345678-1234-5678-9977-123456789111, got %s", *userGroups[0].ID)
	}
}

func (m *MockCrudUserGroupClient) Do(req *http.Request) (io.ReadCloser, error) {
	resp, err := ioutil.ReadFile("./fixtures/crud-usergroup-response.json")
	if err != nil {
		m.T.Fatal(err)
	}

	if req.Method != m.method {
		m.T.Errorf("request expected %s method, got %s", m.method, req.Method)
	}

	body, _ := ioutil.ReadAll(req.Body)

	// The calls for adding/removing users only transmit an array of strings
	// Not an actual UserGroup object.
	var addRemoveBody []string
	if err := json.Unmarshal(body, &addRemoveBody); err == nil {
		return ioutil.NopCloser(bytes.NewReader(resp)), nil
	}

	userGroup := UserGroup{}
	err = json.Unmarshal(body, &userGroup)
	if err != nil {
		m.T.Fatal(err)
	}

	return ioutil.NopCloser(bytes.NewReader(resp)), nil
}


func Test_CreatReadUpdateDelete(t *testing.T) {
	baseurl, _ := url.Parse("http://testing.wavefront.com")
	g := &UserGroups{
		client: &MockCrudUserGroupClient{
			Client: Client{
				Config:     &Config{Token: "1234-5678-9977"},
				BaseURL:    baseurl,
				httpClient: http.DefaultClient,
				debug:      true,
			},
			T: t,
		},
	}

	g.client.(*MockCrudUserGroupClient).method = "POST"

	userGroup := &UserGroup{}
	if err := g.Create(userGroup); err == nil {
		t.Errorf("expected to receive error for missing name")
	}

	userGroup.Name = "testing"
	if err := g.Create(userGroup); err == nil {
		t.Errorf("expected to receive error for missing permissions")
	}

	userGroup.Permissions = []string{ALERTS_MANAGEMENT}
	if err := g.Create(userGroup); err != nil {
		t.Fatal(err)
	}

	if *userGroup.ID != "12345678-1234-5678-9977-123456789111" {
		t.Errorf("expected UserGroup ID of 12345678-1234-5678-9977-123456789111, got %s", *userGroup.ID)
	}

	g.client.(*MockCrudUserGroupClient).method = "PUT"
	var _ = g.Update(userGroup)

	g.client.(*MockCrudUserGroupClient).method = "POST"
	modifyUser := []string{userGroup.Users[0]}

	var _ = g.RemoveUsers(userGroup.ID, &modifyUser)

	var _ = g.AddUsers(userGroup.ID, &modifyUser)

	g.client.(*MockCrudUserGroupClient).method = "DELETE"
	var _ = g.Delete(userGroup)
	if *userGroup.ID != "" {
		t.Errorf("expected UserGroup ID to be blank, got %s", *userGroup.ID)
	}
}