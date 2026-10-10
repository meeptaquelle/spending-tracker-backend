package storage

import (
	"context"
	"encoding/base64"
	"fmt"
)

func NewGoogleSheetsStoreVercel(spreadsheetID, credentialsB64, credentialsPath string) (Store, error) {
	ctx := context.Background()

	if credentialsB64 != "" {
		raw, err := base64.StdEncoding.DecodeString(credentialsB64)
		if err != nil {
			return nil, fmt.Errorf("storage: decode GOOGLE_SERVICE_ACCOUNT_JSON_B64: %w", err)
		}
		return NewGoogleSheetsStoreFromJSON(ctx, spreadsheetID, raw)
	}
	return NewGoogleSheetsStore(ctx, spreadsheetID, credentialsPath)
}
