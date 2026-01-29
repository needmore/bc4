package attachments

import (
	"testing"
)

func TestParseAttachments(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected int
		checkFn  func(*testing.T, []Attachment)
	}{
		{
			name: "single image attachment",
			html: `<bc-attachment sgid="BAh7CEkiCG..." content-type="image/jpeg" width="2560" height="1536" url="https://example.com/image.jpg" href="https://example.com/image.jpg" filename="my-photo.jpg" caption="My photo">
  <figure>
    <img srcset="..." src="...">
    <figcaption>My photo</figcaption>
  </figure>
</bc-attachment>`,
			expected: 1,
			checkFn: func(t *testing.T, attachments []Attachment) {
				if len(attachments) != 1 {
					t.Fatalf("expected 1 attachment, got %d", len(attachments))
				}
				a := attachments[0]
				if a.SGID != "BAh7CEkiCG..." {
					t.Errorf("expected SGID 'BAh7CEkiCG...', got '%s'", a.SGID)
				}
				if a.ContentType != "image/jpeg" {
					t.Errorf("expected content-type 'image/jpeg', got '%s'", a.ContentType)
				}
				if a.Filename != "my-photo.jpg" {
					t.Errorf("expected filename 'my-photo.jpg', got '%s'", a.Filename)
				}
				if a.Caption != "My photo" {
					t.Errorf("expected caption 'My photo', got '%s'", a.Caption)
				}
				if !a.IsImage() {
					t.Error("expected IsImage() to return true")
				}
			},
		},
		{
			name: "multiple attachments",
			html: `<div>
  <bc-attachment presentation="gallery" sgid="SGIDone" content-type="image/png" filename="first.png" url="https://example.com/first.png"></bc-attachment>
  <bc-attachment presentation="gallery" sgid="SGIDtwo" content-type="image/gif" filename="second.gif" url="https://example.com/second.gif"></bc-attachment>
</div>`,
			expected: 2,
			checkFn: func(t *testing.T, attachments []Attachment) {
				if len(attachments) != 2 {
					t.Fatalf("expected 2 attachments, got %d", len(attachments))
				}
				if attachments[0].Filename != "first.png" {
					t.Errorf("expected first filename 'first.png', got '%s'", attachments[0].Filename)
				}
				if attachments[1].Filename != "second.gif" {
					t.Errorf("expected second filename 'second.gif', got '%s'", attachments[1].Filename)
				}
			},
		},
		{
			name:     "self-closing tag",
			html:     `<bc-attachment sgid="TEST123" content-type="application/pdf" filename="document.pdf" url="https://example.com/doc.pdf" />`,
			expected: 1,
			checkFn: func(t *testing.T, attachments []Attachment) {
				if len(attachments) != 1 {
					t.Fatalf("expected 1 attachment, got %d", len(attachments))
				}
				a := attachments[0]
				if a.ContentType != "application/pdf" {
					t.Errorf("expected content-type 'application/pdf', got '%s'", a.ContentType)
				}
				if a.IsImage() {
					t.Error("expected IsImage() to return false for PDF")
				}
			},
		},
		{
			name:     "no attachments",
			html:     `<p>Just some regular HTML content</p>`,
			expected: 0,
			checkFn: func(t *testing.T, attachments []Attachment) {
				if len(attachments) != 0 {
					t.Fatalf("expected 0 attachments, got %d", len(attachments))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachments := ParseAttachments(tt.html)
			if len(attachments) != tt.expected {
				t.Errorf("expected %d attachments, got %d", tt.expected, len(attachments))
			}
			if tt.checkFn != nil {
				tt.checkFn(t, attachments)
			}
		})
	}
}

func TestBuildTag(t *testing.T) {
	tag := BuildTag("SGID123")
	expected := `<bc-attachment sgid="SGID123"></bc-attachment>`
	if tag != expected {
		t.Fatalf("expected %s, got %s", expected, tag)
	}

	if BuildTag("") != "" {
		t.Fatalf("expected empty string when sgid is empty")
	}
}

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name       string
		attachment Attachment
		expected   string
	}{
		{
			name: "with caption",
			attachment: Attachment{
				Caption:  "My Caption",
				Filename: "file.jpg",
			},
			expected: "My Caption",
		},
		{
			name: "without caption, with filename",
			attachment: Attachment{
				Filename: "document.pdf",
			},
			expected: "document.pdf",
		},
		{
			name:       "no caption or filename",
			attachment: Attachment{},
			expected:   "Unnamed attachment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attachment.GetDisplayName()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestIsBlobURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "traditional upload URL",
			url:      "https://3.basecamp.com/1234567/uploads/12345/download/file.jpg",
			expected: false,
		},
		{
			name:     "blob URL with blobs path",
			url:      "https://3.basecamp.com/4446664/blobs/abc123-def456/previews/full?dppx=2",
			expected: true,
		},
		{
			name:     "preview subdomain URL",
			url:      "https://preview.3.basecamp.com/4446664/blobs/abc123/previews/full",
			expected: true,
		},
		{
			name:     "preview subdomain without blobs",
			url:      "https://preview.3.basecamp.com/4446664/some/path",
			expected: true,
		},
		{
			name:     "empty URL",
			url:      "",
			expected: false,
		},
		{
			name:     "regular basecamp URL",
			url:      "https://3.basecamp.com/1234567/buckets/123/messages/456",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBlobURL(tt.url)
			if result != tt.expected {
				t.Errorf("IsBlobURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestExtractUploadIDFromURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		expected  int64
		expectErr bool
	}{
		{
			name:      "valid upload URL with trailing slash",
			url:       "https://3.basecamp.com/1234567/uploads/98765/download/file.jpg",
			expected:  98765,
			expectErr: false,
		},
		{
			name:      "valid upload URL without trailing content",
			url:       "https://3.basecamp.com/1234567/uploads/12345",
			expected:  12345,
			expectErr: false,
		},
		{
			name:      "blob URL should fail",
			url:       "https://preview.3.basecamp.com/4446664/blobs/abc123/previews/full",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "URL without uploads path",
			url:       "https://3.basecamp.com/1234567/buckets/123/messages/456",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "empty URL",
			url:       "",
			expected:  0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractUploadIDFromURL(tt.url)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ExtractUploadIDFromURL(%q) expected error, got nil", tt.url)
				}
			} else {
				if err != nil {
					t.Errorf("ExtractUploadIDFromURL(%q) unexpected error: %v", tt.url, err)
				}
				if result != tt.expected {
					t.Errorf("ExtractUploadIDFromURL(%q) = %d, expected %d", tt.url, result, tt.expected)
				}
			}
		})
	}
}

func TestTryExtractUploadID(t *testing.T) {
	tests := []struct {
		name            string
		attachment      Attachment
		expectErr       bool
		expectBlobURL   bool
		expectedID      int64
		expectedBlobURL string
	}{
		{
			name: "URL has upload ID",
			attachment: Attachment{
				URL:  "https://3.basecamp.com/1234567/uploads/12345/download/file.jpg",
				Href: "https://3.basecamp.com/1234567/uploads/12345/download/file.jpg",
			},
			expectErr:  false,
			expectedID: 12345,
		},
		{
			name: "URL is blob but Href has upload ID",
			attachment: Attachment{
				URL:  "https://preview.3.basecamp.com/4446664/blobs/abc123/previews/full",
				Href: "https://3.basecamp.com/1234567/uploads/67890/download/file.jpg",
			},
			expectErr:  false,
			expectedID: 67890,
		},
		{
			name: "both URL and Href are blob URLs",
			attachment: Attachment{
				URL:  "https://preview.3.basecamp.com/4446664/blobs/abc123/previews/full",
				Href: "https://3.basecamp.com/4446664/blobs/def456/previews/full",
			},
			expectErr:       true,
			expectBlobURL:   true,
			expectedBlobURL: "https://preview.3.basecamp.com/4446664/blobs/abc123/previews/full",
		},
		{
			name: "URL is blob, Href is empty",
			attachment: Attachment{
				URL:  "https://preview.3.basecamp.com/4446664/blobs/abc123/previews/full",
				Href: "",
			},
			expectErr:       true,
			expectBlobURL:   true,
			expectedBlobURL: "https://preview.3.basecamp.com/4446664/blobs/abc123/previews/full",
		},
		{
			name: "URL is empty, Href has upload ID",
			attachment: Attachment{
				URL:  "",
				Href: "https://3.basecamp.com/1234567/uploads/11111/download/file.jpg",
			},
			expectErr:  false,
			expectedID: 11111,
		},
		{
			name: "both URLs invalid, not blob",
			attachment: Attachment{
				URL:  "https://example.com/random/path",
				Href: "https://example.com/other/path",
			},
			expectErr:     true,
			expectBlobURL: false,
		},
		{
			name:       "empty attachment",
			attachment: Attachment{},
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TryExtractUploadID(&tt.attachment)

			if tt.expectErr {
				if err == nil {
					t.Errorf("TryExtractUploadID expected error, got nil")
				}
				if tt.expectBlobURL {
					if result == nil {
						t.Errorf("expected result with blob URL info, got nil")
					} else if !result.IsBlobURL {
						t.Errorf("expected IsBlobURL=true, got false")
					} else if result.BlobURL != tt.expectedBlobURL {
						t.Errorf("expected BlobURL=%q, got %q", tt.expectedBlobURL, result.BlobURL)
					}
				}
			} else {
				if err != nil {
					t.Errorf("TryExtractUploadID unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected non-nil result")
				} else {
					if result.UploadID != tt.expectedID {
						t.Errorf("expected UploadID=%d, got %d", tt.expectedID, result.UploadID)
					}
					if result.IsBlobURL {
						t.Errorf("expected IsBlobURL=false, got true")
					}
				}
			}
		})
	}
}
