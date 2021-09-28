// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"net/url"
	"strings"
)

// ConvertSignatureCDN converts the temporary CDN URL to long period URL.
func ConvertSignatureCDN(cdnURL string) string {
	u, err := url.Parse(cdnURL)
	if err != nil {
		return cdnURL
	}

	u.Host = strings.ReplaceAll(u.Host, "-sign", "")
	u.RawQuery = "" // Clean the signature in query.

	return u.String()
}

// IsGIFImage checks whether the given CDN image file is a GIF image.
func IsGIFImage(cdnURL string) bool {
	req, err := http.NewRequest(http.MethodGet, cdnURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("user-agent", UserAgent)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	
	return resp.Header.Get("Content-Type") == "application/octet-stream"
}
