package drs

// GetObjectResponse ...
type GetObjectResponse struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	SelfURI       string         `json:"self_uri"`
	Size          int64          `json:"size"`
	CreatedTime   string         `json:"created_time"`
	UpdatedTime   string         `json:"updated_time"`
	Version       string         `json:"version"`
	MimeType      string         `json:"mime_type"` // application/json
	Description   string         `json:"description"`
	Aliases       []string       `json:"aliases,omitempty"`
	Checksums     []Checksum     `json:"checksums"`
	Contents      []Content      `json:"contents,omitempty"`
	AccessMethods []AccessMethod `json:"access_methods"`
}

// Checksum ...
type Checksum struct {
	Checksum string `json:"checksum"`
	Type     string `json:"type"`
}

// Content ...
type Content struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	DRSURL   string    `json:"drs_uri"`
	Contents []Content `json:"contents"`
}

// AccessMethod ...
type AccessMethod struct {
	Type           string      `json:"type"` // s3 gs ftp gsiftp globus htsget https file
	AccessURL      AccessURL   `json:"access_url"`
	Region         string      `json:"region"`
	AccessID       string      `json:"access_id"`
	Authorizations interface{} `json:"authorizations,omitempty"`
}

// GetAccessResponse ...
type GetAccessResponse struct {
	AccessURL `json:",inline"`
}

// AccessURL ...
type AccessURL struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}
