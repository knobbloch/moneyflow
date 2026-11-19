package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func BaseURL() string {
	if v := os.Getenv("API_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

func (r Response) Decode(t testing.TB, v any) {
	t.Helper()
	if err := json.Unmarshal(r.Body, v); err != nil {
		t.Fatalf("decode response: %v; body: %s", err, r.Body)
	}
}

func (r Response) ErrorCode() string {
	var e struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(r.Body, &e)
	return e.Error.Code
}

// Send не использует testing.T, поэтому его можно вызывать из горутин.
// body: nil, string (сырое тело) или любое значение (будет сериализовано в JSON).
func Send(method, path string, headers map[string]string, body any) (Response, error) {
	var rdr io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		rdr = strings.NewReader(b)
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			return Response{}, err
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, BaseURL()+path, rdr)
	if err != nil {
		return Response{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return Response{Status: resp.StatusCode, Header: resp.Header, Body: data}, err
}

func Do(t testing.TB, method, path string, headers map[string]string, body any) Response {
	t.Helper()
	r, err := Send(method, path, headers, body)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return r
}

func AssertStatus(t testing.TB, r Response, want int) {
	t.Helper()
	if r.Status != want {
		t.Fatalf("status: got %d, want %d; body: %s", r.Status, want, r.Body)
	}
}

func AssertError(t testing.TB, r Response, status int, code string) {
	t.Helper()
	AssertStatus(t, r, status)
	if got := r.ErrorCode(); got != code {
		t.Fatalf("error code: got %q, want %q; body: %s", got, code, r.Body)
	}
}

// RunMain проверяет, что API запущен, и только потом гоняет тесты.
func RunMain(m *testing.M) int {
	r, err := Send("GET", "/health", nil, nil)
	if err != nil || r.Status != 200 {
		fmt.Fprintf(os.Stderr, "API недоступен по адресу %s (ошибка: %v). Запусти банковское API и повтори.\n", BaseURL(), err)
		return 1
	}
	return m.Run()
}