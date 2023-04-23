package domain

type GetMediaInput struct {
  Name      string `json:"name"`
  Section   string `json:"section"`
  From      string `json:"from"`
  Timestamp int64  `json:"timestamp"`
}

type PutMediaInput struct {
  Name      string `json:"name"`
  Section   string `json:"section"`
  Content   []byte `json:"content"`
  Overwrite bool   `json:"overwrite,omitempty"`
  From      string `json:"from"`
  Timestamp int64  `json:"timestamp"`
}

func (p *PutMediaInput) ToMessage() *PutMessage {
  return &PutMessage{
    MetaInfo: PutMessageMetaInfo{
      Name:      p.Name,
      Section:   p.Section,
      Overwrite: p.Overwrite,
      From:      p.From,
      Timestamp: p.Timestamp,
    },
    Content: p.Content,
  }
}

type Media struct {
  Found   bool   `json:"found"`
  Name    string `json:"name"`
  Section string `json:"section"`
  Path    string `json:"path"`
}
