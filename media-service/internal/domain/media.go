package domain

type GetMediaInput struct {
  Name        string `json:"name"`
  Section     string `json:"section"`
  ContentType string `json:"content_type"`
  From        string `json:"from"`
  Timestamp   int64  `json:"timestamp"`
}

type PutMediaInput struct {
  Name          string `json:"name"`
  Section       string `json:"section"`
  Content       []byte `json:"content"`
  ContentType   string `json:"content_type"`
  ContentLength string `json:"content_length,omitempty"`
  Overwrite     bool   `json:"overwrite,omitempty"`
  From          string `json:"from"`
  Timestamp     int64  `json:"timestamp"`
}

func (p *PutMediaInput) ToMessage() *PutMessage {
  return &PutMessage{
    MetaInfo: PutMessageMetaInfo{
      Name:          p.Name,
      Section:       p.Section,
      ContentType:   p.ContentType,
      ContentLength: p.ContentLength,
      Overwrite:     p.Overwrite,
      From:          p.From,
      Timestamp:     p.Timestamp,
    },
    Content: p.Content,
  }
}

type Media struct {
  SourceUrl string `json:"source_url"`
}
