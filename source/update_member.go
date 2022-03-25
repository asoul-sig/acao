// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package source

import (
	"net/http"

	"github.com/asoul-sig/asoul-video/pkg/model"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	log "unknwon.dev/clog/v2"
)

func init() {
	Register(&UpdateMember{})
}

type UpdateMember struct{}

func (s *UpdateMember) String() string {
	return "update_member"
}

func (s *UpdateMember) Scrap(result chan Result) {
	defer func() { result <- Result{End: true} }()

	for _, secUID := range asoul {
		userInfo, err := scrapMember(secUID)
		if err != nil {
			log.Error("Failed to scrap member data: %v", err)
			continue
		}

		var avatarURL string
		if len(userInfo.UserInfo.AvatarLarger.UrlList) != 0 {
			avatarURL = userInfo.UserInfo.AvatarLarger.UrlList[0]
		} else if len(userInfo.UserInfo.AvatarMedium.UrlList) != 0 {
			avatarURL = userInfo.UserInfo.AvatarMedium.UrlList[0]
		} else if len(userInfo.UserInfo.AvatarThumb.UrlList) != 0 {
			avatarURL = userInfo.UserInfo.AvatarThumb.UrlList[0]
		}

		updateMember := model.UpdateMember{
			SecUID:    secUID,
			UID:       userInfo.UserInfo.Uid,
			UniqueID:  userInfo.UserInfo.UniqueId,
			ShortUID:  userInfo.UserInfo.ShortId,
			Name:      userInfo.UserInfo.Nickname,
			AvatarURL: avatarURL,
			Signature: userInfo.UserInfo.Signature,
		}

		callback, err := jsoniter.Marshal(updateMember)
		if err != nil {
			log.Error("Failed to encode callback JSON: %v", err)
			continue
		}

		log.Trace("Fetch member %q", userInfo.UserInfo.Nickname)

		result <- Result{
			Data: callback,
		}
	}
}

type userInfo struct {
	UserInfo struct {
		AvatarLarger struct {
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
		} `json:"avatar_larger"`
		FollowerCount    int    `json:"follower_count"`
		TotalFavorited   string `json:"total_favorited"`
		CustomVerify     string `json:"custom_verify"`
		Secret           int    `json:"secret"`
		Signature        string `json:"signature"`
		AwemeCount       int    `json:"aweme_count"`
		VerificationType int    `json:"verification_type"`
		OriginalMusician struct {
			MusicCount     int `json:"music_count"`
			MusicUsedCount int `json:"music_used_count"`
		} `json:"original_musician"`
		Region        string      `json:"region"`
		PolicyVersion interface{} `json:"policy_version"`
		ShortId       string      `json:"short_id"`
		Nickname      string      `json:"nickname"`
		AvatarMedium  struct {
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
		} `json:"avatar_medium"`
		FollowingCount   int         `json:"following_count"`
		UniqueId         string      `json:"unique_id"`
		FollowersDetail  interface{} `json:"followers_detail"`
		PlatformSyncInfo interface{} `json:"platform_sync_info"`
		Geofencing       interface{} `json:"geofencing"`
		Uid              string      `json:"uid"`
		TypeLabel        interface{} `json:"type_label"`
		FavoritingCount  int         `json:"favoriting_count"`
		IsGovMediaVip    bool        `json:"is_gov_media_vip"`
		AvatarThumb      struct {
			Uri     string   `json:"uri"`
			UrlList []string `json:"url_list"`
		} `json:"avatar_thumb"`
	} `json:"user_info"`
}

func scrapMember(secUID model.MemberSecUID) (*userInfo, error) {
	respBody, err := SimpleScrap(http.MethodGet, "https://www.iesdouyin.com/web/api/v2/user/info/?sec_uid="+string(secUID))
	if err != nil {
		return nil, errors.Wrap(err, "scrap")
	}

	var userInfo userInfo
	if err := jsoniter.Unmarshal(respBody, &userInfo); err != nil {
		return nil, errors.Wrap(err, "JSON decode")
	}
	return &userInfo, nil
}
