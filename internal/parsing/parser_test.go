package parsing

import (
	"testing"
)

func TestApacheParser_CommonFormat(t *testing.T) {
	p := &ApacheParser{}
	raw := `127.0.0.1 - frank [10/Oct/2026:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`
	result, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}
	if result.Fields["client_ip"] != "127.0.0.1" {
		t.Errorf("Expected client_ip 127.0.0.1, got %v", result.Fields["client_ip"])
	}
	if result.Fields["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", result.Fields["method"])
	}
	if result.Fields["status"] != 200 {
		t.Errorf("Expected status 200, got %v", result.Fields["status"])
	}
	if result.Fields["bytes"] != 2326 {
		t.Errorf("Expected bytes 2326, got %v", result.Fields["bytes"])
	}
	if result.Timestamp == nil {
		t.Error("Expected non-nil timestamp")
	}
}

func TestApacheParser_CombinedFormat(t *testing.T) {
	p := &ApacheParser{}
	raw := `127.0.0.1 - james [09/Oct/2026:22:10:18 +0000] "POST /api/submit HTTP/1.1" 201 534 "https://example.com" "Mozilla/5.0 (X11; Linux x86_64)"`
	result, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}
	if result.Fields["referrer"] != "https://example.com" {
		t.Errorf("Expected referrer, got %v", result.Fields["referrer"])
	}
	if result.Fields["user_agent"] != "Mozilla/5.0 (X11; Linux x86_64)" {
		t.Errorf("Expected user_agent, got %v", result.Fields["user_agent"])
	}
}

func TestApacheParser_UnsupportedLayout(t *testing.T) {
	p := &ApacheParser{}
	raw := `This is not an Apache log at all`
	_, err := p.Parse(raw)
	if err == nil {
		t.Fatal("Expected UNSUPPORTED_LAYOUT error")
	}
}

func TestApacheParser_Empty(t *testing.T) {
	p := &ApacheParser{}
	_, err := p.Parse("")
	if err == nil {
		t.Fatal("Expected error for empty input")
	}
}

func TestNginxParser_CombinedFormat(t *testing.T) {
	p := &NginxParser{}
	raw := `192.168.1.1 - - [10/Oct/2026:13:55:36 -0700] "GET /index.html HTTP/1.1" 200 612 "-" "Mozilla/5.0"`
	result, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}
	if result.Fields["remote_addr"] != "192.168.1.1" {
		t.Errorf("Expected remote_addr 192.168.1.1, got %v", result.Fields["remote_addr"])
	}
	if result.Fields["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", result.Fields["method"])
	}
	if result.Fields["status"] != 200 {
		t.Errorf("Expected status 200, got %v", result.Fields["status"])
	}
}

func TestNginxParser_UnsupportedLayout(t *testing.T) {
	p := &NginxParser{}
	_, err := p.Parse("not a nginx log")
	if err == nil {
		t.Fatal("Expected UNSUPPORTED_LAYOUT error")
	}
}

func TestJSONParser_ValidJSON(t *testing.T) {
	p := &JSONParser{}
	raw := `{"timestamp": "2026-10-10T14:30:00Z", "action": "login", "user": "admin", "result": "success"}`
	result, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}
	if result.Fields["action"] != "login" {
		t.Errorf("Expected action login, got %v", result.Fields["action"])
	}
	if result.Fields["user"] != "admin" {
		t.Errorf("Expected user admin, got %v", result.Fields["user"])
	}
	if result.Timestamp == nil {
		t.Error("Expected non-nil timestamp")
	}
}

func TestJSONParser_MalformedJSON(t *testing.T) {
	p := &JSONParser{}
	_, err := p.Parse(`{bad json}`)
	if err == nil {
		t.Fatal("Expected PARSE_ERROR for malformed JSON")
	}
}

func TestJSONParser_Empty(t *testing.T) {
	p := &JSONParser{}
	_, err := p.Parse("")
	if err == nil {
		t.Fatal("Expected error for empty input")
	}
}

func TestSyslogParser_RFC3164(t *testing.T) {
	p := &SyslogParser{}
	raw := `<34>Oct 11 22:14:15 mymachine sshd[12345]: Failed password for user admin from 10.0.0.1 port 22 ssh2`
	result, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}
	if result.Fields["hostname"] != "mymachine" {
		t.Errorf("Expected hostname mymachine, got %v", result.Fields["hostname"])
	}
	if result.Fields["app_name"] != "sshd" {
		t.Errorf("Expected app_name sshd, got %v", result.Fields["app_name"])
	}
	if result.Fields["pid"] != 12345 {
		t.Errorf("Expected pid 12345, got %v", result.Fields["pid"])
	}
	if result.Fields["priority"] != 34 {
		t.Errorf("Expected priority 34, got %v", result.Fields["priority"])
	}
}

func TestSyslogParser_NoPID(t *testing.T) {
	p := &SyslogParser{}
	raw := `<13>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick`
	result, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Expected successful parse, got error: %v", err)
	}
	if _, ok := result.Fields["pid"]; ok {
		t.Error("Expected no pid field")
	}
}

func TestSyslogParser_Malformed(t *testing.T) {
	p := &SyslogParser{}
	_, err := p.Parse("not a syslog message at all")
	if err == nil {
		t.Fatal("Expected error for malformed syslog")
	}
}
