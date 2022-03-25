// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package source

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/asoul-sig/asoul-video/pkg/model"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	log "unknwon.dev/clog/v2"

	"github.com/asoul-sig/acao/util"
)

func init() {
	Register(&UpdateVideoMeta{})
}

type UpdateVideoMeta struct{}

func (s *UpdateVideoMeta) String() string {
	return "update_video_meta"
}

func (s *UpdateVideoMeta) Scrap(result chan Result) {
	defer func() { result <- Result{End: true} }()

	page := 0
	for {
		page++

		videoMeta, err := s.scrapVideoList(page)
		if err != nil {
			if err == ErrNoMoreVideos {
				break
			}
		}

		callback, err := jsoniter.Marshal(videoMeta)
		if err != nil {
			log.Error("Failed to encode callback JSON: %v", err)
			continue
		}

		result <- Result{
			Data: callback,
		}
	}
}

type video struct {
	ID               string    `json:"id"`
	VID              string    `json:"vid"`
	OriginCoverURLs  []string  `json:"origin_cover_urls"`
	DynamicCoverURLs []string  `json:"dynamic_cover_urls"`
	CreatedAt        time.Time `json:"created_at"`
}

var ErrNoMoreVideos = errors.New("no more videos")

func (s *UpdateVideoMeta) scrapVideoList(page int) ([]*model.UpdateVideoMeta, error) {
	resp, err := http.Get("https://asoul.cdn.n3ko.co/api/videos?page=" + strconv.Itoa(page))
	if err != nil {
		return nil, errors.Wrap(err, "get video list from asoul-video")
	}
	defer func() { _ = resp.Body.Close() }()

	var respJSON struct {
		Data []video `json:"data"`
	}
	if err := jsoniter.NewDecoder(resp.Body).Decode(&respJSON); err != nil {
		log.Error("Failed to decode JSON: %v", err)
		return nil, errors.Wrap(err, "decode JSON")
	}
	if len(respJSON.Data) == 0 {
		return nil, ErrNoMoreVideos
	}

	updateVideoMetas := make([]*model.UpdateVideoMeta, 0, len(respJSON.Data))
	for _, video := range respJSON.Data {
		id := video.ID

		var metaData *videoMeta
		var err error
		for i := 1; i <= 3; i++ {
			metaData, err = s.getVideoMeta(id)
			if err != nil || len(metaData.ItemList) == 0 {
				log.Warn("Failed to get video %q meta data [ %d / 3 ]: %v", id, i, err)
				continue
			}
			break
		}
		if err != nil {
			log.Error("Failed to get video %q meta data: %v", err)
			continue
		}

		if len(metaData.ItemList) == 0 {
			log.Error("Video %q not found", id)
			continue
		}

		meta := metaData.ItemList[0]
		createdAt := time.Unix(int64(meta.CreateTime), 0)

		originCoverURLs := make([]string, 0, len(meta.Video.OriginCover.UrlList))
		for _, url := range meta.Video.OriginCover.UrlList {
			originCoverURLs = append(originCoverURLs, util.ConvertSignatureCDN(url))
		}

		dynamicCoverURLs := make([]string, 0, len(meta.Video.DynamicCover.UrlList))
		for _, url := range meta.Video.DynamicCover.UrlList {
			dynamicCoverURLs = append(dynamicCoverURLs, util.ConvertSignatureCDN(url))
		}

		isDynamicCover := len(dynamicCoverURLs) > 0 && util.IsGIFImage(dynamicCoverURLs[0])

		updateVideoMetas = append(updateVideoMetas, &model.UpdateVideoMeta{
			ID:               id,
			VID:              meta.Video.Vid,
			OriginCoverURLs:  originCoverURLs,
			DynamicCoverURLs: dynamicCoverURLs,
			IsDynamicCover:   isDynamicCover,
			CreatedAt:        createdAt,

			Statistic: model.Statistic{
				Share:   meta.Statistics.ShareCount,
				Forward: meta.Statistics.ForwardCount,
				Digg:    meta.Statistics.DiggCount,
				Play:    meta.Statistics.PlayCount,
				Comment: meta.Statistics.CommentCount,
			},
		})
	}

	return updateVideoMetas, nil
}

type videoMeta struct {
	StatusCode int `json:"status_code"`
	ItemList   []struct {
		IsLiveReplay bool `json:"is_live_replay"`
		TextExtra    []struct {
			End         int    `json:"end"`
			Type        int    `json:"type"`
			HashtagName string `json:"hashtag_name"`
			HashtagId   int64  `json:"hashtag_id"`
			Start       int    `json:"start"`
		} `json:"text_extra"`
		AuthorUserId int64       `json:"author_user_id"`
		LongVideo    interface{} `json:"long_video"`
		Images       interface{} `json:"images"`
		ChaList      []struct {
			ConnectMusic   interface{} `json:"connect_music"`
			Type           int         `json:"type"`
			ViewCount      int         `json:"view_count"`
			HashTagProfile string      `json:"hash_tag_profile"`
			Cid            string      `json:"cid"`
			Desc           string      `json:"desc"`
			IsCommerce     bool        `json:"is_commerce"`
			ChaName        string      `json:"cha_name"`
			UserCount      int         `json:"user_count"`
		} `json:"cha_list"`
		Statistics struct {
			DiggCount    int64  `json:"digg_count"`
			PlayCount    int64  `json:"play_count"`
			ShareCount   int64  `json:"share_count"`
			AwemeId      string `json:"aweme_id"`
			CommentCount int64  `json:"comment_count"`
			ForwardCount int64  `json:"forward_count"`
		} `json:"statistics"`
		RiskInfos struct {
			Warn    bool   `json:"warn"`
			Type    int    `json:"type"`
			Content string `json:"content"`
		} `json:"risk_infos"`
		Desc  string `json:"desc"`
		Music struct {
			Mid         string `json:"mid"`
			CoverMedium struct {
				UrlList []string `json:"url_list"`
				Uri     string   `json:"uri"`
			} `json:"cover_medium"`
			CoverThumb struct {
				UrlList []string `json:"url_list"`
				Uri     string   `json:"uri"`
			} `json:"cover_thumb"`
			Duration int         `json:"duration"`
			Position interface{} `json:"position"`
			Id       int64       `json:"id"`
			Author   string      `json:"author"`
			CoverHd  struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"cover_hd"`
			CoverLarge struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"cover_large"`
			PlayUrl struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"play_url"`
			Status int    `json:"status"`
			Title  string `json:"title"`
		} `json:"music"`
		Video struct {
			PlayAddr struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"play_addr"`
			DynamicCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"dynamic_cover"`
			BitRate      interface{} `json:"bit_rate"`
			Vid          string      `json:"vid"`
			Ratio        string      `json:"ratio"`
			HasWatermark bool        `json:"has_watermark"`
			Duration     int         `json:"duration"`
			Cover        struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"cover"`
			Height      int `json:"height"`
			Width       int `json:"width"`
			OriginCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"origin_cover"`
		} `json:"video"`
		AwemeType    int         `json:"aweme_type"`
		VideoText    interface{} `json:"video_text"`
		GroupId      int64       `json:"group_id"`
		LabelTopText interface{} `json:"label_top_text"`
		IsPreview    int         `json:"is_preview"`
		Author       struct {
			PlatformSyncInfo interface{} `json:"platform_sync_info"`
			Geofencing       interface{} `json:"geofencing"`
			PolicyVersion    interface{} `json:"policy_version"`
			ShortId          string      `json:"short_id"`
			Nickname         string      `json:"nickname"`
			AvatarMedium     struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"avatar_medium"`
			UniqueId        string      `json:"unique_id"`
			FollowersDetail interface{} `json:"followers_detail"`
			TypeLabel       interface{} `json:"type_label"`
			Uid             string      `json:"uid"`
			Signature       string      `json:"signature"`
			AvatarLarger    struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"avatar_larger"`
			AvatarThumb struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"avatar_thumb"`
		} `json:"author"`
		ForwardId   string      `json:"forward_id"`
		CreateTime  int         `json:"create_time"`
		VideoLabels interface{} `json:"video_labels"`
		ImageInfos  interface{} `json:"image_infos"`
		Duration    int         `json:"duration"`
		CommentList interface{} `json:"comment_list"`
		Geofencing  interface{} `json:"geofencing"`
		AwemeId     string      `json:"aweme_id"`
		ShareUrl    string      `json:"share_url"`
		ShareInfo   struct {
			ShareWeiboDesc string `json:"share_weibo_desc"`
			ShareDesc      string `json:"share_desc"`
			ShareTitle     string `json:"share_title"`
		} `json:"share_info"`
		Promotions interface{} `json:"promotions"`
	} `json:"item_list"`
	Extra struct {
		Now   int64  `json:"now"`
		Logid string `json:"logid"`
	} `json:"extra"`
}

func (s *UpdateVideoMeta) getVideoMeta(id string) (*videoMeta, error) {
	time.Sleep(500 * time.Millisecond)

	signature := util.MakeSignature("e99p1ant", util.UserAgent)
	log.Trace("Signature: %v for video: %q", signature, id)

	url := fmt.Sprintf("https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids=%s&_signature=%s", id, signature)
	respBody, err := SimpleScrap(http.MethodGet, url)
	if err != nil {
		return nil, errors.Wrap(err, "scrap")
	}

	var videoMeta videoMeta
	if err := jsoniter.Unmarshal(respBody, &videoMeta); err != nil {
		return nil, errors.Wrap(err, "decode JSON")
	}
	return &videoMeta, nil
}
