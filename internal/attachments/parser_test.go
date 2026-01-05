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
