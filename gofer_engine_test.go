package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNormalizeGopherInputSeparatesItemTypeAndSelector(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType byte
		wantSel  string
	}{
		{
			name:     "text file",
			input:    "gopher://example.com/0docs/readme",
			wantType: '0',
			wantSel:  "docs/readme",
		},
		{
			name:     "menu with slash selector",
			input:    "gopher://example.com/1/docs",
			wantType: '1',
			wantSel:  "/docs",
		},
		{
			name:     "root menu",
			input:    "gopher://example.com/",
			wantType: '1',
			wantSel:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeGopherInput(tt.input)
			if err != nil {
				t.Fatalf("normalizeGopherInput() error = %v", err)
			}
			if got.Type != tt.wantType || got.Selector != tt.wantSel {
				t.Errorf("normalizeGopherInput() = type %q, selector %q; want type %q, selector %q",
					got.Type, got.Selector, tt.wantType, tt.wantSel)
			}
		})
	}
}

func TestNormalizeGopherInputDropsUnknownItemTypeAndDefaultsToText(t *testing.T) {
	got, err := normalizeGopherInput("foo.bar:70/users/iroll/gofer-project")
	if err != nil {
		t.Fatalf("normalizeGopherInput() error = %v", err)
	}
	if got.Type != '0' || got.Selector != "sers/iroll/gofer-project" {
		t.Errorf("normalizeGopherInput() = type %q, selector %q; want type %q, selector %q",
			got.Type, got.Selector, '0', "sers/iroll/gofer-project")
	}
	if !got.BadItemType {
		t.Error("normalizeGopherInput() BadItemType = false, want true")
	}
}

func TestLocalGopherURLPreservesItemType(t *testing.T) {
	target := GopherTarget{
		Host:        "example.com",
		Port:        "70",
		Selector:    "docs/readme",
		Type:        '0',
		BadItemType: true,
	}

	got, err := url.Parse(localGopherURL(target))
	if err != nil {
		t.Fatalf("url.Parse(localGopherURL()) error = %v", err)
	}

	query := got.Query()
	if query.Get("type") != "0" {
		t.Errorf("type query = %q, want %q", query.Get("type"), "0")
	}
	if query.Get("selector") != target.Selector {
		t.Errorf("selector query = %q, want %q", query.Get("selector"), target.Selector)
	}
	if query.Get("badtype") != "1" {
		t.Errorf("badtype query = %q, want %q", query.Get("badtype"), "1")
	}
}

func TestBadItemTypeWarningRendersBeforeServerError(t *testing.T) {
	serverMenu := "3Not found\t\tnull.host\t1\n.\n"
	menu := prependBadItemTypeWarning(serverMenu)
	page := formatMenuHTML(menu, "example.com", "70", "users", '0', false)

	warningIconIndex := strings.Index(page, "[ERR]")
	warningTextIndex := strings.Index(page, "Bad or missing item type")
	serverErrorIndex := strings.Index(page, "Not found")
	if warningIconIndex == -1 || warningTextIndex < warningIconIndex ||
		serverErrorIndex == -1 || warningTextIndex > serverErrorIndex {
		t.Errorf("rendered page = %q; want bad-item-type warning before server response", page)
	}
}

func TestFormatMenuHTMLIncludesItemTypeInTitleAndInput(t *testing.T) {
	page := formatMenuHTML("", "sdf.org", "70", "users/iroll/gofer-project", '1', false)

	for _, want := range []string{
		"<title>gofer - sdf.org:70/1users/iroll/gofer-project</title>",
		`value="sdf.org:70/1users/iroll/gofer-project"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("rendered page does not contain %q", want)
		}
	}
}

func TestServeGopherRendersUnknownItemTypeAsText(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	requestLine := make(chan string, 1)
	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()

		line, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			serverErr <- err
			return
		}
		requestLine <- line
		_, err = io.WriteString(conn, "page content\n")
		serverErr <- err
	}()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	gopherURI := fmt.Sprintf("gopher://127.0.0.1:%s/uusers", port)
	request := httptest.NewRequest(http.MethodGet, "/?uri="+url.QueryEscape(gopherURI), nil)
	response := httptest.NewRecorder()

	serveGopher(response, request)

	if got, want := <-requestLine, "users\r\n"; got != want {
		t.Errorf("Gopher request = %q, want %q", got, want)
	}
	if got := response.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", got)
	}
	if got, want := response.Body.String(), "[ERR] Bad or missing item type\npage content\n"; got != want {
		t.Errorf("response body = %q, want %q", got, want)
	}
	if err := <-serverErr; err != nil {
		t.Errorf("Gopher test server error = %v", err)
	}
}
