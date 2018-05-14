package anchore

import (
	"encoding/json"
	"testing"
)

func TestGetRequest(t *testing.T) {
	api := api{
		username: "admin",
		password: "foobar",
		endpoint: "http://echo.jsontest.com/foo/bar/data/1234",
	}
	dat := make(map[string]interface{}, 2)
	req, err := api.newGetRequest(api.endpoint)
	if err != nil {
		t.Error(err)
	}
	resp, err := request(req)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&dat)
	if err != nil {
		t.Error(err)
	}

	if dat["data"] != "1234" {
		t.Errorf("Parsed value (data) seems to be wrong. Value: %s", dat["data"])
	} else if dat["foo"] != "bar" {
		t.Errorf("Parsed value (foo) seems to be wrong. Value: %s", dat["foo"])
	}
}

func TestPostRequest(t *testing.T) {
	api := api{
		username: "admin",
		password: "foobar",
		endpoint: "http://validate.jsontest.com",
	}
	dat := make(map[string]interface{}, 2)
	dat["foo"] = "bar"
	dat["data"] = "1234"

	req, err := api.newPostRequest(api.endpoint, &dat)
	if err != nil {
		t.Error(err)
	}
	_, err = request(req)
	if err != nil {
		t.Error(err)
	}
}
