package domain

import "encoding/json"

type PutMessage struct {
  MetaInfo *PutMessageMetaInfo `json:"meta_info"`
  Content  []byte              `json:"content"`
}

type PutMessageMetaInfo struct {
  Name          string `json:"name"`
  Section       string `json:"section"`
  ContentType   string `json:"content_type"`
  ContentLength string `json:"content_length,omitempty"`
  Overwrite     bool   `json:"overwrite,omitempty"`
  From          string `json:"from"`
  Timestamp     int64  `json:"timestamp"`
}

func (m *PutMessageMetaInfo) MarshalMap() (map[string]any, error) {
  b, err := json.Marshal(m)
  if err != nil {
    return nil, err
  }
  mMap := map[string]any{}
  if err := json.Unmarshal(b, &mMap); err != nil {
    return nil, err
  }
  return mMap, nil
}
