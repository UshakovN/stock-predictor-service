package polygon

import (
	"fmt"
	"main/internal/domain"
	"strings"

	"github.com/UshakovN/stock-predictor-service/utils"
)

func (f *Fetcher) formMsgForBrandingImage(tickerId, imageURL, brandingType string) (*domain.PutMessage, error) {
	const (
		sectionName = "polygon_references"
		nameDashSep = "-"
		nameDotSep  = "."
	)
	imageResp, err := f.client.GetFullResp(imageURL)
	if err != nil {
		return nil, fmt.Errorf("cannot get image response for ticker '%s': %v", tickerId, err)
	}

	imageExtension, err := utils.ExtractFileExtension(imageURL)
	if err != nil {
		return nil, fmt.Errorf("cannot extract image extension from url '%s': %v", err, imageURL)
	}

	// ticker_id-branding_type.extension
	imageName := fmt.Sprint(tickerId, nameDashSep, brandingType, nameDotSep, imageExtension)

	return &domain.PutMessage{
		MetaInfo: &domain.PutMessageMetaInfo{
			Name:      imageName,
			Section:   sectionName,
			Overwrite: false,
			From:      fetcherName,
			Timestamp: utils.NowTimestampUTC(),
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
