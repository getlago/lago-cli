package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestStableAndBeta(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/releases/latest" {
			_, _ = io.WriteString(response, `{"tag_name":"v1.2.0","html_url":"https://example.test/release"}`)
			return
		}
		_, _ = io.WriteString(response, `[{"tag_name":"v1.3.0-beta.2","prerelease":true},{"tag_name":"v1.3.0-beta.1","prerelease":true}]`)
	}))
	defer server.Close()
	check, _, err := Latest(context.Background(), "1.1.0", "stable", "test", server.URL)
	if err != nil || !check.UpdateAvailable || check.Latest != "1.2.0" {
		t.Fatalf("check=%#v error=%v", check, err)
	}
	beta, _, err := Latest(context.Background(), "1.2.0", "beta", "test", server.URL)
	if err != nil || !beta.UpdateAvailable || beta.Latest != "1.3.0-beta.2" {
		t.Fatalf("beta=%#v error=%v", beta, err)
	}
}

func TestExtractTarBinaryAndChecksum(t *testing.T) {
	t.Parallel()
	var archive bytes.Buffer
	gzipWriter := gzip.NewWriter(&archive)
	tarWriter := tar.NewWriter(gzipWriter)
	content := []byte("fake-binary")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "lago", Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	_, _ = tarWriter.Write(content)
	_ = tarWriter.Close()
	_ = gzipWriter.Close()
	extracted, err := extractBinary(archive.Bytes(), "tar.gz")
	if err != nil || !bytes.Equal(extracted, content) {
		t.Fatalf("extracted=%q error=%v", extracted, err)
	}
	if got := checksumFor([]byte("abc  lago_1.0.0_linux_amd64.tar.gz\n"), "lago_1.0.0_linux_amd64.tar.gz"); got != "abc" {
		t.Fatalf("checksum = %q", got)
	}
}
