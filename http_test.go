package directadmin

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"strings"
	"testing"
)

// The framing is assembled by hand around a reader, so it has to parse back as a normal
// multipart form, and the length it reports has to be exact: it becomes Content-Length, and
// a wrong one either truncates the upload or hangs waiting for bytes that never come.
func TestMultipartFileBody(t *testing.T) {
	const content = "hello world"

	body, length, contentType, err := multipartFileBody("sqlfile", "dump.sql", strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	raw, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}

	if int64(len(raw)) != length {
		t.Fatalf("reported length %d, wrote %d", length, len(raw))
	}

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatal(err)
	}

	reader := multipart.NewReader(bytes.NewReader(raw), params["boundary"])

	part, err := reader.NextPart()
	if err != nil {
		t.Fatal(err)
	}

	if part.FormName() != "sqlfile" {
		t.Errorf("form name = %q, want sqlfile", part.FormName())
	}

	if part.FileName() != "dump.sql" {
		t.Errorf("file name = %q, want dump.sql", part.FileName())
	}

	got, err := io.ReadAll(part)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != content {
		t.Errorf("content = %q, want %q", got, content)
	}

	if _, err = reader.NextPart(); err != io.EOF {
		t.Errorf("expected exactly one part, got %v", err)
	}
}

// The size no longer comes from the caller, so the inference is what keeps Content-Length
// honest for each body type an upload actually gets handed.
func TestReaderSize(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "upload")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = file.WriteString("hello"); err != nil {
		t.Fatal(err)
	}

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name    string
		content io.Reader
		want    int64
	}{
		{"bytes.Reader", bytes.NewReader([]byte("hello")), 5},
		{"bytes.Buffer", bytes.NewBufferString("hello"), 5},
		{"strings.Reader", strings.NewReader("hello"), 5},
		{"os.File", file, 5},
	} {
		got, err := readerSize(test.content)
		if err != nil {
			t.Errorf("%s: %v", test.name, err)
		} else if got != test.want {
			t.Errorf("%s: size = %d, want %d", test.name, got, test.want)
		}
	}

	if _, err = readerSize(io.LimitReader(file, 5)); err == nil {
		t.Error("expected an error for a reader of unknown length")
	}

	// A file the caller has already read from reports what is left, not its size.
	if _, err = file.Seek(2, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	got, err := readerSize(file)
	if err != nil {
		t.Fatal(err)
	}

	if got != 3 {
		t.Errorf("size of a partially read file = %d, want 3", got)
	}
}

// The peek has to be invisible to a caller: a dump longer than the peek limit must come
// back whole, and an error document must not be handed over as if it were data.
func TestCheckStreamError(t *testing.T) {
	dump := "-- MySQL dump\n" + strings.Repeat("INSERT INTO t VALUES (1);\n", 200)

	if len(dump) <= peekLimit {
		t.Fatalf("test dump is %d bytes, needs to exceed the %d byte peek", len(dump), peekLimit)
	}

	stream, err := checkStreamError(io.NopCloser(strings.NewReader(dump)))
	if err != nil {
		t.Fatal(err)
	}

	got, err := io.ReadAll(stream)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != dump {
		t.Errorf("read %d bytes, want %d", len(got), len(dump))
	}

	for _, body := range []string{
		`{"error":"Cannot download file","result":"File does not exist"}`,
		"\n  {\"error\":\"truncated",
		"{not json at all",
	} {
		if _, err = checkStreamError(io.NopCloser(strings.NewReader(body))); err == nil {
			t.Errorf("%s: expected an error", body)
		}
	}
}
