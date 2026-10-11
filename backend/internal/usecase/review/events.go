package review

type ReviewCreatedEvent struct {
	ReviewID    int64  `json:"review_id"`
	AuthorID    int32  `json:"author_id"`
	TargetID    int32  `json:"target_id"`
	TargetType  int32  `json:"target_type"`
	Rating      int32  `json:"rating"`
	Comment     string `json:"comment"`
	ReferenceID string `json:"reference_id"`
}

type ReviewRepliedEvent struct {
	ReviewID int64  `json:"review_id"`
	AuthorID int32  `json:"author_id"`
	TargetID int32  `json:"target_id"`
	Reply    string `json:"reply"`
}
