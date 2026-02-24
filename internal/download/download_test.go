package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/needmore/bc4/internal/api"
)

// mockUploadOps implements api.UploadOperations for testing.
type mockUploadOps struct {
	uploads         map[int64]*api.Upload // uploadID -> Upload
	getUploadError  error                 // global error for all GetUpload calls
	downloadError   error                 // global error for all DownloadAttachment calls
	downloadedPaths []string              // tracks paths passed to DownloadAttachment
	getUploadCalls  []int64               // tracks upload IDs requested
}

func (m *mockUploadOps) GetUpload(_ context.Context, _ string, uploadID int64) (*api.Upload, error) {
	m.getUploadCalls = append(m.getUploadCalls, uploadID)
	if m.getUploadError != nil {
		return nil, m.getUploadError
	}
	u, ok := m.uploads[uploadID]
	if !ok {
		return nil, fmt.Errorf("upload %d not found", uploadID)
	}
	return u, nil
}

func (m *mockUploadOps) DownloadAttachment(_ context.Context, _ string, destPath string) error {
	m.downloadedPaths = append(m.downloadedPaths, destPath)
	if m.downloadError != nil {
		return m.downloadError
	}
	// Create the file to simulate a real download
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(destPath, []byte("test content"), 0644)
}

// htmlWithUploadAttachment returns HTML containing a bc-attachment with an upload URL.
func htmlWithUploadAttachment(uploadID int64, filename string) string {
	return fmt.Sprintf(
		`<bc-attachment sgid="test" content-type="image/png" filename="%s" url="https://3.basecamp.com/123/uploads/%d/download/%s"></bc-attachment>`,
		filename, uploadID, filename,
	)
}

// htmlWithBlobAttachment returns HTML containing a bc-attachment with a blob URL.
func htmlWithBlobAttachment() string {
	return `<bc-attachment sgid="blob" content-type="image/jpeg" filename="photo.jpg" url="https://3.basecamp.com/123/blobs/abc123/previews/full"></bc-attachment>`
}

func TestDownloadFromSources_NoAttachments(t *testing.T) {
	mock := &mockUploadOps{}
	sources := []AttachmentSource{
		{Label: "card", Content: "<p>No attachments here</p>"},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected Total=0, got %d", result.Total)
	}
	if result.Successful != 0 || result.Failed != 0 || result.Skipped != 0 {
		t.Errorf("expected all counts zero, got successful=%d failed=%d skipped=%d",
			result.Successful, result.Failed, result.Skipped)
	}
}

func TestDownloadFromSources_EmptySources(t *testing.T) {
	mock := &mockUploadOps{}
	sources := []AttachmentSource{}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected Total=0, got %d", result.Total)
	}
}

func TestDownloadFromSources_SingleSourceSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			100: {ID: 100, Filename: "report.pdf", ByteSize: 2048, DownloadURL: "https://example.com/dl/100"},
		},
	}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "report.pdf")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected Total=1, got %d", result.Total)
	}
	if result.Successful != 1 {
		t.Errorf("expected Successful=1, got %d", result.Successful)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}

	// Verify file was created
	destPath := filepath.Join(tmpDir, "report.pdf")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", destPath)
	}

	// Verify the mock was called correctly
	if len(mock.getUploadCalls) != 1 || mock.getUploadCalls[0] != 100 {
		t.Errorf("expected GetUpload called with ID 100, got %v", mock.getUploadCalls)
	}
}

func TestDownloadFromSources_MultipleSourcesSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			100: {ID: 100, Filename: "card-image.png", ByteSize: 1024, DownloadURL: "https://example.com/dl/100"},
			200: {ID: 200, Filename: "comment-file.pdf", ByteSize: 4096, DownloadURL: "https://example.com/dl/200"},
		},
	}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "card-image.png")},
		{Label: "comment #5 by Alice", Content: htmlWithUploadAttachment(200, "comment-file.pdf")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected Total=2, got %d", result.Total)
	}
	if result.Successful != 2 {
		t.Errorf("expected Successful=2, got %d", result.Successful)
	}

	// Verify both files were created
	for _, name := range []string{"card-image.png", "comment-file.pdf"} {
		if _, err := os.Stat(filepath.Join(tmpDir, name)); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", name)
		}
	}
}

func TestDownloadFromSources_AttachmentIndexValid(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			200: {ID: 200, Filename: "second.png", ByteSize: 512, DownloadURL: "https://example.com/dl/200"},
		},
	}
	// Two attachments in content, but we only want the second
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "first.png") + htmlWithUploadAttachment(200, "second.png")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir:       tmpDir,
		AttachmentIndex: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected Total=1, got %d", result.Total)
	}
	if result.Successful != 1 {
		t.Errorf("expected Successful=1, got %d", result.Successful)
	}

	// Only second file should exist
	if _, err := os.Stat(filepath.Join(tmpDir, "second.png")); os.IsNotExist(err) {
		t.Error("expected second.png to exist")
	}
	// GetUpload should only have been called for upload 200
	if len(mock.getUploadCalls) != 1 || mock.getUploadCalls[0] != 200 {
		t.Errorf("expected only GetUpload(200), got %v", mock.getUploadCalls)
	}
}

func TestDownloadFromSources_AttachmentIndexOutOfRange(t *testing.T) {
	mock := &mockUploadOps{}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "only.png")},
	}

	_, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		AttachmentIndex: 5,
	})
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
	if want := "attachment index 5 out of range (found 1 attachments)"; err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}

func TestDownloadFromSources_GetUploadError(t *testing.T) {
	mock := &mockUploadOps{
		getUploadError: fmt.Errorf("API rate limited"),
	}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "file.png")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{})
	if err == nil {
		t.Fatal("expected error when all downloads fail")
	}
	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
	if result.Successful != 0 {
		t.Errorf("expected Successful=0, got %d", result.Successful)
	}
}

func TestDownloadFromSources_DownloadAttachmentError(t *testing.T) {
	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			100: {ID: 100, Filename: "file.png", ByteSize: 1024, DownloadURL: "https://example.com/dl/100"},
		},
		downloadError: fmt.Errorf("network timeout"),
	}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "file.png")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error when download fails")
	}
	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
}

func TestDownloadFromSources_FileExistsSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create the file
	existingFile := filepath.Join(tmpDir, "existing.png")
	if err := os.WriteFile(existingFile, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			100: {ID: 100, Filename: "existing.png", ByteSize: 1024, DownloadURL: "https://example.com/dl/100"},
		},
	}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "existing.png")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir: tmpDir,
		Overwrite: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("expected Skipped=1, got %d", result.Skipped)
	}
	if result.Successful != 0 {
		t.Errorf("expected Successful=0, got %d", result.Successful)
	}

	// Verify original file was not overwritten
	content, _ := os.ReadFile(existingFile)
	if string(content) != "original" {
		t.Errorf("file was overwritten, expected 'original', got %q", string(content))
	}

	// Verify DownloadAttachment was never called
	if len(mock.downloadedPaths) != 0 {
		t.Errorf("expected no download calls, got %v", mock.downloadedPaths)
	}
}

func TestDownloadFromSources_FileExistsOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create the file
	existingFile := filepath.Join(tmpDir, "existing.png")
	if err := os.WriteFile(existingFile, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			100: {ID: 100, Filename: "existing.png", ByteSize: 1024, DownloadURL: "https://example.com/dl/100"},
		},
	}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "existing.png")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir: tmpDir,
		Overwrite: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Successful != 1 {
		t.Errorf("expected Successful=1, got %d", result.Successful)
	}
	if result.Skipped != 0 {
		t.Errorf("expected Skipped=0, got %d", result.Skipped)
	}

	// Verify file was overwritten
	content, _ := os.ReadFile(existingFile)
	if string(content) == "original" {
		t.Error("file was not overwritten")
	}
}

func TestDownloadFromSources_BlobURLHandled(t *testing.T) {
	mock := &mockUploadOps{}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithBlobAttachment()},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{})
	if err == nil {
		t.Fatal("expected error when all downloads fail (blob URL)")
	}
	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
	// GetUpload should not have been called for blob URLs
	if len(mock.getUploadCalls) != 0 {
		t.Errorf("GetUpload should not be called for blob URLs, got %v", mock.getUploadCalls)
	}
}

func TestDownloadFromSources_MixedSuccessAndFailure(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			100: {ID: 100, Filename: "good.png", ByteSize: 1024, DownloadURL: "https://example.com/dl/100"},
			// upload 200 is missing from the map, so GetUpload will fail for it
		},
	}

	// Two upload attachments: one succeeds, one fails (upload not found)
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "good.png") + htmlWithUploadAttachment(200, "missing.png")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir: tmpDir,
	})
	if err == nil {
		t.Fatal("expected error when some downloads fail")
	}
	if result.Successful != 1 {
		t.Errorf("expected Successful=1, got %d", result.Successful)
	}
	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
	if result.Total != 2 {
		t.Errorf("expected Total=2, got %d", result.Total)
	}
}

func TestDownloadFromSources_DefaultOutputDir(t *testing.T) {
	// When OutputDir is empty, it should default to "."
	// We can test this indirectly by checking the downloaded path
	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			100: {ID: 100, Filename: "file.png", ByteSize: 512, DownloadURL: "https://example.com/dl/100"},
		},
	}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "file.png")},
	}

	// Use a temp dir as working directory to avoid polluting the repo
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir: "", // should default to "."
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Successful != 1 {
		t.Errorf("expected Successful=1, got %d", result.Successful)
	}
	// File should be in "." which is tmpDir
	if _, err := os.Stat(filepath.Join(tmpDir, "file.png")); os.IsNotExist(err) {
		t.Error("expected file.png in current directory")
	}
}

func TestDownloadFromSources_AttachmentIndexAcrossMultipleSources(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			300: {ID: 300, Filename: "from-comment.png", ByteSize: 256, DownloadURL: "https://example.com/dl/300"},
		},
	}
	// Source 1 has 1 attachment (index 1), source 2 has 1 attachment (index 2)
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "from-card.png")},
		{Label: "comment #5", Content: htmlWithUploadAttachment(300, "from-comment.png")},
	}

	// Request attachment index 2 -- should be the comment attachment
	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir:       tmpDir,
		AttachmentIndex: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Successful != 1 {
		t.Errorf("expected Successful=1, got %d", result.Successful)
	}
	// Should have downloaded from-comment.png, not from-card.png
	if len(mock.getUploadCalls) != 1 || mock.getUploadCalls[0] != 300 {
		t.Errorf("expected GetUpload(300), got %v", mock.getUploadCalls)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "from-comment.png")); os.IsNotExist(err) {
		t.Error("expected from-comment.png to exist")
	}
}

func TestDownloadFromSources_DuplicateFilenames(t *testing.T) {
	tmpDir := t.TempDir()
	mock := &mockUploadOps{
		uploads: map[int64]*api.Upload{
			100: {ID: 100, Filename: "image.png", ByteSize: 1024, DownloadURL: "https://example.com/dl/100"},
			200: {ID: 200, Filename: "image.png", ByteSize: 2048, DownloadURL: "https://example.com/dl/200"},
		},
	}
	sources := []AttachmentSource{
		{Label: "card", Content: htmlWithUploadAttachment(100, "image.png")},
		{Label: "comment #1 by Bob", Content: htmlWithUploadAttachment(200, "image.png")},
	}

	result, err := DownloadFromSources(context.Background(), mock, "bucket1", sources, Options{
		OutputDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Successful != 2 {
		t.Errorf("expected Successful=2, got %d", result.Successful)
	}

	// Both files should exist with different names
	if _, err := os.Stat(filepath.Join(tmpDir, "image.png")); os.IsNotExist(err) {
		t.Error("expected image.png to exist")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "image_1.png")); os.IsNotExist(err) {
		t.Error("expected image_1.png to exist (deduplicated)")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal filename", "photo.jpg", "photo.jpg"},
		{"path traversal", "../../../etc/passwd", "passwd"},
		{"path separators", "path/to/file.txt", "file.txt"},
		{"control characters", "file\x00name.txt", "filename.txt"},
		{"unsafe characters", "file<>:\"|?*.txt", "file_______.txt"},
		{"empty string", "", "attachment"},
		{"dot only", ".", "attachment"},
		{"double dot", "..", "attachment"},
		{"spaces preserved", "my file.txt", "my file.txt"},
		{"unicode preserved", "日本語.pdf", "日本語.pdf"},
		{"windows reserved CON", "CON", "_CON"},
		{"windows reserved CON.txt", "CON.txt", "_CON.txt"},
		{"windows reserved PRN", "PRN", "_PRN"},
		{"windows reserved NUL", "NUL", "_NUL"},
		{"windows reserved COM1", "COM1", "_COM1"},
		{"windows reserved LPT1.log", "LPT1.log", "_LPT1.log"},
		{"windows reserved lowercase", "con.txt", "_con.txt"},
		{"not reserved COM0", "COM0", "COM0"},
		{"not reserved CONNECT", "CONNECT.pdf", "CONNECT.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatByteSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"small bytes", 500, "500 B"},
		{"one KB", 1024, "1.0 KB"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"large megabytes", 5242880, "5.0 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatByteSize(tt.input)
			if result != tt.expected {
				t.Errorf("FormatByteSize(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
