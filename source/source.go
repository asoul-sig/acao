// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package source

import (
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const userAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36`

var Sources = make(map[string]Source)

type Source interface {
	String() string
	Scrap() ([]jsoniter.RawMessage, error)
}

func Register(source Source) {
	Sources[source.String()] = source
}

func SimpleScrap(method, url string) (jsoniter.RawMessage, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "request")
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}

	if resp.StatusCode/100 != 2 {
		return nil, errors.Errorf("unexpected status code %d: %q", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}
