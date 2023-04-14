package polygon

import (
  "fmt"
  "main/internal/domain"
  "main/pkg/utils"
  "strings"
)

func (f *Fetcher) formMsgForBrandingImage(tickerId, imageURL, brandingType string) (*domain.PutMessage, error) {
  const (
    headerContentType   = "Content-Type"
    headerContentLength = "Content-Length"
    blobContentType     = "application/octet-stream"
    sectionName         = "polygon_references"
    nameDashSep         = "-"
  )
  imageResp, err := f.client.GetFullResp(imageURL)
  if err != nil {
    return nil, fmt.Errorf("cannot get image response for ticker '%s': %v", tickerId, err)
  }

  contentType, ok := imageResp.Headers.Get(headerContentType)
  if !ok {
    contentType = blobContentType
  }
  imageName := fmt.Sprint(tickerId, nameDashSep, brandingType)
  contentLength := imageResp.Headers.GetOrDefault(headerContentLength)

  return &domain.PutMessage{
    MetaInfo: &domain.PutMessageMetaInfo{
      Name:          imageName,
      Section:       sectionName,
      ContentType:   contentType,
      ContentLength: contentLength,
      Overwrite:     false,
      From:          fetcherName,
      Timestamp:     utils.NowTimestampUTC(),
    },
    Content: imageResp.Content,
  }, nil
}

func (f *Fetcher) sendMessagesToPutTickerBranding(tickerId string, branding *tickerDetailsBranding) error {
  const (
    brandingTypeIcon = "icon"
    brandingTypeLogo = "logo"
  )
  if tickerId == "" || branding == nil {
    return nil
  }
  iconURL := strings.TrimSpace(branding.IconUrl)
  if iconURL != "" {
    // form and send message if icon url not empty
    iconPutMsg, err := f.formMsgForBrandingImage(tickerId, iconURL, brandingTypeIcon)
    if err != nil {
      return err
    }
    if err = f.msQueue.PublishMessage(iconPutMsg); err != nil {
      return err
    }
  }

  logoURL := strings.TrimSpace(branding.LogoUrl)
  if logoURL != "" {
    // form and send message if logo url not empty
    logoPutMsg, err := f.formMsgForBrandingImage(tickerId, logoURL, brandingTypeLogo)
    if err != nil {
      return err
    }
    if err = f.msQueue.PublishMessage(logoPutMsg); err != nil {
      return err
    }
  }

  return nil
}
