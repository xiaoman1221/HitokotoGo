package entity

type V struct {
	ProtocolVersion string `json:"protocol_version"`
	BundleVersion   string `json:"bundle_version"`
	UpdatedAt       int64  `json:"updated_at"`
	Categories      struct {
		Path      string `json:"path"`
		Timestamp int64  `json:"timestamp"`
	} `json:"categories"`
	Sentences []struct {
		Name      string `json:"name"`
		Key       string `json:"key"`
		Path      string `json:"path"`
		Timestamp int64  `json:"timestamp"`
	} `json:"sentences"`
}
