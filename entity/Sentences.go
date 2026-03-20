package entity

type SentencesSimple struct {
	Id           *int    `json:"id"`
	Uuid         *string `json:"uuid"`
	Hitokoto     *string `json:"hitokoto"`
	SentenceType *string `json:"type"`
	From         *string `json:"from"`
	FromWho      *string `json:"from_who"`
	Creator      *string `json:"creator"`
	CreatorUid   *int    `json:"creator_uid"`
	Reviewer     *int    `json:"reviewer"`
	CommitFrom   *string `json:"commit_from"`
	CreatedAt    *string `json:"created_at"`
	Length       *int    `json:"length"`
}
type SentencesVersion struct {
	ProtocolVersion string                     `json:"protocol_version"`
	BundleVersion   string                     `json:"bundle_version"`
	UpDateAt        *string                    `json:"update_at"`
	Categories      SentencesVersionCategories `json:"categories"`
	Sentences       []SentencesCategories      `json:"sentences"`
}
type SentencesCategories struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	Desc     string  `json:"desc"`
	Key      string  `json:"key"`
	CreateAt *string `json:"create_at"`
	UpdateAt *string `json:"update_at"`
	Path     string  `json:"path"`
}
type SentencesVersionCategories struct {
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
}
