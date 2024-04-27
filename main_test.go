package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Tax struct {
	Tax float64 `json:"tax"`
}

func TestPostEndpoint(t *testing.T) {
	// arrange
	body := bytes.NewBufferString(`{
	  "totalIncome": 500000.0,
	  "wht": 0.0,
	  "allowances": [
		{
		  "allowanceType": "donation",
		  "amount": 0.0
		}
	  ]
	}`)
	var u Tax

	res := request(http.MethodPost, uri("tax/calculations"), body)
	err := res.Decode(&u)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "29000", u.Tax)
}

func uri(paths ...string) string {
	host := "http://localhost:8080"
	if paths == nil {
		return host
	}

	url := append([]string{host}, paths...)
	return strings.Join(url, "/")
}

func request(method, url string, body io.Reader) *Response {
	req, _ := http.NewRequest(method, url, body)
	token := os.Getenv("AUTH_TOKEN")
	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	return &Response{res, err}
}

type Response struct {
	*http.Response
	err error
}

func (r *Response) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}

	return json.NewDecoder(r.Body).Decode(v)
}
