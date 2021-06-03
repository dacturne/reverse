package reverse

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestScanner_Scan_success(t *testing.T) {
	expected := []string{"line one", "line two", "last line"}
	payload := strings.Join(expected, "\n")
	rd := strings.NewReader(payload)

	out := make([]string, 0, len(expected))
	s := NewScanner(rd, int64(len(payload)))
	s.SetChunkSize(512)
	s.SetDelimiter('\n')

	for i := 0; i < len(expected); i++ {
		line, err := s.Scan()
		if err != nil {
			t.FailNow()
		}
		out = append(out, string(line))
	}

	if _, err := s.Scan(); err == nil {
		if !errors.Is(err, io.EOF) {
			t.Fatalf("expected an EOF but got: %v", err)
		}
	}

	for i := range expected {
		x := len(expected) - 1 - i
		t.Logf("Comparing '%s' with '%s'\n", expected[x], out[i])
		if expected[x] != out[i] {
			t.FailNow()
		}
	}
}

func TestScanner_Scan_ErrBufferTooSmall(t *testing.T) {
	payload := "example payload"
	rd := strings.NewReader(payload)
	s := NewScanner(rd, int64(len(payload)))
	s.SetMaxBufferSize(1)

	_, err := s.Scan()
	if err == nil || !errors.Is(err, ErrBufferTooSmall) {
		t.Fatalf("expected an ErrBufferTooSmall but got: %v", err)
	}
}
