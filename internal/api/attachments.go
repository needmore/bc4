package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// AttachmentUploadResponse represents the response from the attachments endpoint.
type AttachmentUploadResponse struct {
	AttachableSGID string `json:"attachable_sgid"`
	Filename       string `json:"filename"`
	ContentType    string `json:"content_type"`
	ByteSize       int64  `json:"byte_size"`
}

// UploadAttachment uploads raw file data to Basecamp and returns the attachable SGID.
// contentType is optional; if empty, http.DetectContentType is used as a fallback.
func (c *Client) UploadAttachment(filename string, data []byte, contentType string) (*AttachmentUploadResponse, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	ct := contentType
	if ct == "" {
		ct = http.DetectContentType(data)
	}

	path := fmt.Sprintf("/attachments.json?name=%s", url.QueryEscape(filename))
	resp, err := c.doRequestWithHeaders("POST", path, bytes.NewReader(data), map[string]string{
		"Content-Type": ct,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var upload AttachmentUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
		return nil, fmt.Errorf("failed to decode attachment response: %w", err)
	}

	return &upload, nil
}
