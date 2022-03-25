// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package source

import (
	"io"
	"net/http"

	"github.com/asoul-sig/asoul-video/pkg/model"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"

	"github.com/asoul-sig/acao/util"
)

var asoul = []model.MemberSecUID{
	model.MemberSecUIDAva,
	model.MemberSecUIDBella,
	model.MemberSecUIDCarol,
	model.MemberSecUIDDiana,
	model.MemberSecUIDEileen,
	model.MemberSecUIDAcao,
}

type Result struct {
	Data jsoniter.RawMessage
	End  bool
}

var Sources = make(map[string]Source)

type Source interface {
	String() string
	Scrap(result chan Result)
}

func Register(source Source) {
	Sources[source.String()] = source
}

func SimpleScrap(method, url string) (jsoniter.RawMessage, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}
	req.Header.Set("User-Agent", util.UserAgent)

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
