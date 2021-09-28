// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package source

import (
	"fmt"
	"net/http"

	"github.com/asoul-video/asoul-video/pkg/model"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	log "unknwon.dev/clog/v2"

	"github.com/asoul-video/acao/util"
)

func init() {
	Register(&CreateVideo{})
}

type CreateVideo struct{}

func (s *CreateVideo) String() string {
	return "create_video"
}

func (s *CreateVideo) Scrap(result chan Result) {
	defer func() { result <- Result{End: true} }()

	for _, secUID := range asoul {
		cursor := int64(0)

		for {
			memberVideos, nextCursor, err := scrapMemberVideos(secUID, cursor)
			if err != nil {
				log.Error("Failed to scrap member videos: %v", err)
				continue
			}

			for _, video := range memberVideos {
				log.Trace("Fetch video %q", video.Description)
			}

			callback, err := jsoniter.Marshal(memberVideos)
			if err != nil {
				log.Error("Failed to encode callback JSON: %v", err)
				continue
			}
			result <- Result{
				Data: callback,
			}

			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}
	}
}

type videoInfo struct {
	AwemeList []struct {
		ChaList      interface{} `json:"cha_list"`
		ImageInfos   interface{} `json:"image_infos"`
		CommentList  interface{} `json:"comment_list"`
		Geofencing   interface{} `json:"geofencing"`
		LabelTopText interface{} `json:"label_top_text"`
		Images       interface{} `json:"images"`
		Author       struct {
			UniqueId            string `json:"unique_id"`
			WithCommerceEntry   bool   `json:"with_commerce_entry"`
			Nickname            string `json:"nickname"`
			FavoritingCount     int    `json:"favoriting_count"`
			WithFusionShopEntry bool   `json:"with_fusion_shop_entry"`
			AvatarLarger        struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"avatar_larger"`
			AwemeCount     int         `json:"aweme_count"`
			IsAdFake       bool        `json:"is_ad_fake"`
			Signature      string      `json:"signature"`
			FollowerCount  int         `json:"follower_count"`
			Region         string      `json:"region"`
			SecUid         string      `json:"sec_uid"`
			TotalFavorited string      `json:"total_favorited"`
			CustomVerify   string      `json:"custom_verify"`
			PolicyVersion  interface{} `json:"policy_version"`
			UserCanceled   bool        `json:"user_canceled"`
			TypeLabel      []int64     `json:"type_label"`
			Uid            string      `json:"uid"`
			AvatarMedium   struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"avatar_medium"`
			EnterpriseVerifyReason string      `json:"enterprise_verify_reason"`
			PlatformSyncInfo       interface{} `json:"platform_sync_info"`
			HasOrders              bool        `json:"has_orders"`
			VideoIcon              struct {
				Uri     string        `json:"uri"`
				UrlList []interface{} `json:"url_list"`
			} `json:"video_icon"`
			ShortId        string      `json:"short_id"`
			FollowStatus   int         `json:"follow_status"`
			FollowingCount int         `json:"following_count"`
			WithShopEntry  bool        `json:"with_shop_entry"`
			Secret         int         `json:"secret"`
			Geofencing     interface{} `json:"geofencing"`
			AvatarThumb    struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"avatar_thumb"`
			StoryOpen        bool        `json:"story_open"`
			VerificationType int         `json:"verification_type"`
			FollowersDetail  interface{} `json:"followers_detail"`
			IsGovMediaVip    bool        `json:"is_gov_media_vip"`
			Rate             int         `json:"rate"`
		} `json:"author"`
		TextExtra []struct {
			Start       int    `json:"start"`
			End         int    `json:"end"`
			Type        int    `json:"type"`
			HashtagName string `json:"hashtag_name"`
			HashtagId   int64  `json:"hashtag_id"`
			UserId      string `json:"user_id,omitempty"`
		} `json:"text_extra"`
		VideoLabels interface{} `json:"video_labels"`
		AwemeType   int         `json:"aweme_type"`
		VideoText   interface{} `json:"video_text"`
		LongVideo   interface{} `json:"long_video"`
		AwemeId     string      `json:"aweme_id"`
		Video       struct {
			Ratio        string `json:"ratio"`
			DownloadAddr struct {
				UrlList []string `json:"url_list"`
				Uri     string   `json:"uri"`
			} `json:"download_addr"`
			PlayAddrLowbr struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"play_addr_lowbr"`
			BitRate      interface{} `json:"bit_rate"`
			Duration     int64       `json:"duration"`
			Width        int         `json:"width"`
			DynamicCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"dynamic_cover"`
			Height      int `json:"height"`
			OriginCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"origin_cover"`
			HasWatermark bool   `json:"has_watermark"`
			Vid          string `json:"vid"`
			PlayAddr     struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"play_addr"`
			Cover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"cover"`
		} `json:"video"`
		Promotions interface{} `json:"promotions"`
		Desc       string      `json:"desc"`
		Statistics struct {
			PlayCount    int64  `json:"play_count"`
			ShareCount   int64  `json:"share_count"`
			ForwardCount int64  `json:"forward_count"`
			AwemeId      string `json:"aweme_id"`
			CommentCount int64  `json:"comment_count"`
			DiggCount    int64  `json:"digg_count"`
		} `json:"statistics"`
	} `json:"aweme_list"`
	MaxCursor int64 `json:"max_cursor"`
	MinCursor int64 `json:"min_cursor"`
	HasMore   bool  `json:"has_more"`
}

func scrapMemberVideos(secUID model.MemberSecUID, cursor int64) (videos []*model.CreateVideo, nextCursor int64, _ error) {
	signature := util.MakeSignature("e99p1ant", userAgent)
	log.Trace("Signature: %v", signature)

	url := fmt.Sprintf("https://www.iesdouyin.com/web/api/v2/aweme/post/?sec_uid=%s&count=50&max_cursor=%d&_signature=%s", secUID, cursor, signature)

	respBody, err := SimpleScrap(http.MethodGet, url)
	if err != nil {
		return nil, 0, errors.Wrap(err, "scrap")
	}

	var videoInfo videoInfo
	if err := jsoniter.Unmarshal(respBody, &videoInfo); err != nil {
		return nil, 0, errors.Wrap(err, "decode JSON")
	}

	createVideos := make([]*model.CreateVideo, 0, len(videoInfo.AwemeList))
	for _, video := range videoInfo.AwemeList {
		textExtra := make([]string, 0, len(video.TextExtra))
		for _, extra := range video.TextExtra {
			textExtra = append(textExtra, extra.HashtagName)
		}

		createVideos = append(createVideos, &model.CreateVideo{
			ID:               video.AwemeId,
			VID:              video.Video.Vid,
			AuthorSecUID:     model.MemberSecUID(video.Author.SecUid),
			Description:      video.Desc,
			TextExtra:        textExtra,
			OriginCoverURLs:  video.Video.OriginCover.UrlList,
			DynamicCoverURLs: video.Video.DynamicCover.UrlList,
			VideoHeight:      video.Video.Height,
			VideoWidth:       video.Video.Width,
			VideoDuration:    video.Video.Duration,
			VideoRatio:       video.Video.Ratio,
			VideoURLs:        video.Video.PlayAddr.UrlList,
			VideoCDNURL:      "", // TODO Upload to my CDN.

			Statistic: model.Statistic{
				Share:   video.Statistics.ShareCount,
				Forward: video.Statistics.ForwardCount,
				Digg:    video.Statistics.DiggCount,
				Play:    video.Statistics.PlayCount,
				Comment: video.Statistics.CommentCount,
			},
		})
	}

	return createVideos, videoInfo.MaxCursor, nil
}
