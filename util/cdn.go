// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"net/url"
	"strings"

	jsoniter "github.com/json-iterator/go"
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

type imageInfo struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
	Size   int    `json:"size"`
	Md5    string `json:"md5"`
}

// IsGIFImage checks whether the given CDN image file is a GIF image.
func IsGIFImage(cdnURL string) bool {
	cdnURL = strings.ReplaceAll(cdnURL, "/cdn/", "/")
	infoURL := strings.SplitN(cdnURL, "~", 1)[0] + "~info"

	req, err := http.NewRequest(http.MethodGet, infoURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("user-agent", UserAgent)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return false
	}

	var info imageInfo
	if err := jsoniter.NewDecoder(resp.Body).Decode(&info); err != nil {
		return false
	}
	return info.Format == "gif" || info.Format == "webp"
}
