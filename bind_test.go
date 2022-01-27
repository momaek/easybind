package easyreq

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Status string

type queryUsersArgs struct {
	IDs    []Status `in:"query:ids"`
	Status *Status  `in:"query:status"`
	Age    int      `json:"age"`
	OK     bool     `json:"ok"`
}

type testCase struct {
	Query    map[string][]string
	Path     map[string]string
	JSONBody map[string]interface{}
	FormBody map[string][]string
}

func TestBind(t *testing.T) {
	queries := url.Values{}
	queries.Add("ids", "1")
	queries.Add("ids", "3")
	queries.Add("ids", "10086")
	queries.Set("status", "active")
	body := ` {
		"time": "2021-22",
		"age":  20,
		"ok":   true}`
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://hello.world/users?%s", queries.Encode()), strings.NewReader(body))

	args := queryUsersArgs{}
	err := Bind(req, &args)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(args.IDs))
	assert.Equal(t, Status("active"), *args.Status)
	assert.Equal(t, 20, args.Age)
	assert.Equal(t, true, args.OK)
	fmt.Printf("===== %#v \n", args)
}
