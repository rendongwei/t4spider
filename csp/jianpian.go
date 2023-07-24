package csp

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/patrickmn/go-cache"
	"jinadam.github.io/t4spider/common"
)

var (
	classes      = `[{"type_id":"2","type_name":"电视剧"},{"type_id":"1","type_name":"电影"},{"type_id":"4","type_name":"综艺"},{"type_id":"3","type_name":"动漫"}]`
	filterstring = `{"1":[{"key":"year","name":"年代","value":[{"n":"全部","v":"0"},{"n":"2023","v":"153"},{"n":"2022","v":"101"},{"n":"2021","v":"118"},{"n":"2020","v":"16"},{"n":"2019","v":"7"},{"n":"2018","v":"2"},{"n":"2017","v":"3"},{"n":"2016","v":"22"}]},{"key":"sort","name":"排序","value":[{"n":"热门","v":"hot"},{"n":"评分","v":"rating"},{"n":"更新","v":"update"}]}],"2":[{"key":"year","name":"年代","value":[{"n":"全部","v":"0"},{"n":"2023","v":"153"},{"n":"2022","v":"101"},{"n":"2021","v":"118"},{"n":"2020","v":"16"},{"n":"2019","v":"7"},{"n":"2018","v":"2"},{"n":"2017","v":"3"},{"n":"2016","v":"22"}]},{"key":"sort","name":"排序","value":[{"n":"热门","v":"hot"},{"n":"评分","v":"rating"},{"n":"更新","v":"update"}]}],"3":[{"key":"year","name":"年代","value":[{"n":"全部","v":"0"},{"n":"2023","v":"153"},{"n":"2022","v":"101"},{"n":"2021","v":"118"},{"n":"2020","v":"16"},{"n":"2019","v":"7"},{"n":"2018","v":"2"},{"n":"2017","v":"3"},{"n":"2016","v":"22"}]},{"key":"sort","name":"排序","value":[{"n":"热门","v":"hot"},{"n":"评分","v":"rating"},{"n":"更新","v":"update"}]}],"4":[{"key":"year","name":"年代","value":[{"n":"全部","v":"0"},{"n":"2023","v":"153"},{"n":"2022","v":"101"},{"n":"2021","v":"118"},{"n":"2020","v":"16"},{"n":"2019","v":"7"},{"n":"2018","v":"2"},{"n":"2017","v":"3"},{"n":"2016","v":"22"}]},{"key":"sort","name":"排序","value":[{"n":"热门","v":"hot"},{"n":"评分","v":"rating"},{"n":"更新","v":"update"}]}]}`
	picSuffix    = `@Referer=www.jianpianapp.com@User-Agent=jianpian-version353`
	c            *cache.Cache
)

func init() {
	c = cache.New(5*time.Minute, 10*time.Minute)
}

type JianPianSpider struct{}

func (spider *JianPianSpider) client() *req.Client {
	r := req.C().SetBaseURL("http://api2.rinhome.com").
		SetUserAgent("jianpian-android/350").
		SetCommonHeader("JPAUTH", "y261ow7kF2dtzlxh1GS9EB8nbTxNmaK/QQIAjctlKiEv")
	return r
}

func (spider *JianPianSpider) Home() common.HomeResponse {
	if v, ok := c.Get("jianpianhome"); ok {
		return v.(common.HomeResponse)
	}
	var classobj []common.Class
	json.Unmarshal([]byte(classes), &classobj)
	var filterobj map[string]interface{}
	json.Unmarshal([]byte(filterstring), &filterobj)
	var response JianPianHomeApiResponse
	res, err := req.C().R().SetHeaders(map[string]string{
		"User-Agent": "jianpian-android/350",
		"JPAUTH":     "y261ow7kF2dtzlxh1GS9EB8nbTxNmaK/QQIAjctlKiEv",
	}).SetSuccessResult(&response).Get("http://yjpapipxblwdohpakljwg.hxhzs.com/api/tag/hand?code=unknown601193cf375db73d&channel=wandoujia")
	if res.IsSuccessState() {
		if err != nil {
			return common.HomeResponse{}
		}
		var vods []common.Vod
		for _, v := range response.Data {
			for _, vv := range v.Video {
				vods = append(vods, common.Vod{
					VodID:      strconv.Itoa(vv.ID),
					VodName:    vv.Title,
					VodPic:     vv.Path + picSuffix,
					VodRemarks: vv.Playlist.Title + " | " + vv.Score + "分",
				})
			}
		}
		homeResponse := common.HomeResponse{
			List:   vods,
			Filter: filterobj,
			Class:  classobj,
		}
		c.Set("jianpianhome", homeResponse, cache.DefaultExpiration)
		return homeResponse
	}
	return common.HomeResponse{}
}
func (spider *JianPianSpider) Cate(tid string, pg int, ext string) common.CateResponse {
	b, err := base64.StdEncoding.DecodeString(ext)
	if err != nil {
		return common.CateResponse{}
	}
	var extobj map[string]string
	json.Unmarshal(b, &extobj)
	queryParams := map[string]any{
		"area":        0,
		"category_id": tid,
		"page":        pg,
		"type":        0,
		"limit":       24,
		"year":        extobj["year"],
		"sort":        "hot",
	}
	var response JianPianCateApiResponse
	r, err := spider.client().R().SetQueryParamsAnyType(queryParams).SetSuccessResult(&response).Get("/api/crumb/list")
	if err != nil || r.IsErrorState() {
		return common.CateResponse{}
	}
	var vods []common.Vod
	for _, v := range response.Data {
		vods = append(vods, common.Vod{
			VodID:      strconv.Itoa(v.ID),
			VodName:    v.Title,
			VodPic:     v.Path + picSuffix,
			VodRemarks: v.Playlist.Title + " | " + v.Score + "分",
		})
	}
	pgcount := pg
	if len(vods) == 24 {
		pgcount = pg + 1
	}
	return common.CateResponse{
		List:      vods,
		Pagecount: int64(pgcount),
		Page:      pg,
		Total:     math.MaxInt32,
		Limit:     24,
	}
}
func (spider *JianPianSpider) Detail(id string) common.DetailResponse {
	if v, ok := c.Get("jianpiandetail" + id); ok {
		return v.(common.DetailResponse)
	}
	var response JianPianDetailApiResponse
	r, err := spider.client().R().SetQueryParamsAnyType(map[string]any{
		"channel": "wandoujia",
		"token":   "",
		"id":      id,
	}).SetSuccessResult(&response).Get("/api/node/detail")
	if err != nil || r.IsErrorState() {
		return common.DetailResponse{}
	}

	var types []string
	for _, v := range response.Data.Types {
		types = append(types, v.Name)
	}
	var actors []string
	for _, v := range response.Data.Actors {
		actors = append(actors, v.Name)
	}
	var directors []string
	for _, v := range response.Data.Directors {
		directors = append(directors, v.Name)
	}
	var urls []string
	for _, v := range response.Data.BtboDownlist {
		urls = append(urls, v.Val)
	}
	var vods []common.Vod
	vods = append(vods, common.Vod{
		VodID:       strconv.Itoa(response.Data.ID),
		VodName:     response.Data.Title,
		VodPic:      response.Data.Thumbnail + picSuffix,
		VodRemarks:  "豆瓣ID：" + strconv.Itoa(response.Data.DoubanID) + " | " + response.Data.Score + "分",
		TypeName:    strings.Join(types, ","),
		VodActor:    strings.Join(actors, ","),
		VodDirector: strings.Join(directors, ","),
		VodYear:     response.Data.Year.Title,
		VodArea:     response.Data.Area.Title,
		VodContent:  response.Data.Description,
		VodPlayFrom: "七夜-荐片",
		VodPlayURL:  strings.Join(urls, "$$$"),
	})
	detailResponse := common.DetailResponse{
		List: vods,
	}
	c.Set("jianpiandetail"+id, detailResponse, cache.DefaultExpiration)
	return detailResponse
}
func (spider *JianPianSpider) Play(id string) common.PlayResponse {
	return common.PlayResponse{
		URL:   `tvbox-xg:` + id,
		Parse: "0",
	}
}
func (spider *JianPianSpider) Search(key string, flag bool) common.SearchResponse {
	queryParams := map[string]any{
		"key":  key,
		"page": 1,
	}
	var response JianPianSearchApiResponse
	r, err := spider.client().R().SetQueryParamsAnyType(queryParams).SetSuccessResult(&response).Get("/api/video/search")
	if err != nil || r.IsErrorState() {
		return common.SearchResponse{}
	}
	var vods []common.Vod
	for _, v := range response.Data {
		vods = append(vods, common.Vod{
			VodID:      strconv.Itoa(v.ID),
			VodName:    v.Title,
			VodPic:     v.Thumbnail + picSuffix,
			VodRemarks: v.Mask + " | " + v.Score + "分",
		})
	}
	return common.SearchResponse{
		List: vods,
	}
}

type JianPianHomeApiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Title string `json:"title"`
		Video []struct {
			ID             int    `json:"id"`
			Title          string `json:"title"`
			Score          string `json:"score"`
			Finished       int    `json:"finished"`
			Shared         int    `json:"shared"`
			IsLook         int    `json:"is_look"`
			Standbytime    int    `json:"standbytime"`
			Definition     int    `json:"definition"`
			PlaylistLength int    `json:"playlist_length"`
			Playlist       struct {
				Title string `json:"title"`
			} `json:"playlist"`
			Path string `json:"path"`
		} `json:"video"`
	} `json:"data"`
}

type JianPianCateApiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ID             int    `json:"id"`
		Title          string `json:"title"`
		Score          string `json:"score"`
		Finished       int    `json:"finished"`
		Shared         int    `json:"shared"`
		IsLook         int    `json:"is_look"`
		Standbytime    int    `json:"standbytime"`
		Definition     int    `json:"definition"`
		EpisodesCount  string `json:"episodes_count"`
		PlaylistLength int    `json:"playlist_length"`
		Playlist       struct {
			Title string `json:"title"`
		} `json:"playlist"`
		Path   string `json:"path"`
		Tvimg  string `json:"tvimg"`
		CateID int    `json:"cate_id"`
	} `json:"data"`
}

type JianPianDetailApiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ID           int    `json:"id"`
		Title        string `json:"title"`
		Score        string `json:"score"`
		Description  string `json:"description"`
		DoubanID     int    `json:"douban_id"`
		OriginalName string `json:"original_name"`
		OthersName   []struct {
			Value string `json:"value"`
		} `json:"others_name"`
		Duration        string `json:"duration"`
		EpisodesCount   string `json:"episodes_count"`
		EpisodeDuration string `json:"episode_duration"`
		ImdbURL         string `json:"imdb_url"`
		Year            struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"year"`
		Area struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"area"`
		Shared      int    `json:"shared"`
		IsLook      int    `json:"is_look"`
		Standbytime int    `json:"standbytime"`
		Definition  int    `json:"definition"`
		UpdateCycle string `json:"update_cycle"`
		Changed     int    `json:"changed"`
		Finished    int    `json:"finished"`
		Thumbnail   string `json:"thumbnail"`
		Tvimg       string `json:"tvimg"`
		Languages   []struct {
			Value string `json:"value"`
		} `json:"languages"`
		ReleaseDates []struct {
			Value string `json:"value"`
		} `json:"release_dates"`
		Types []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"types"`
		Tags []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"tags"`
		Directors []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"directors"`
		Writers []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"writers"`
		Actors []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"actors"`
		Category []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"category"`
		PlaylistShared int `json:"playlist_shared"`
		HaveM3U8Ur     int `json:"have_m3u8_ur"`
		HaveFtpUr      int `json:"have_ftp_ur"`
		BtboDownlist   []struct {
			ID        int    `json:"id"`
			Val       string `json:"val"`
			Title     string `json:"title"`
			URL       string `json:"url"`
			NeedShare int    `json:"need_share"`
			M3U8URL   string `json:"m3u8_url"`
		} `json:"btbo_downlist"`
		XunleiDownlist []interface{} `json:"xunlei_downlist"`
		M3U8Downlist   []struct {
			ID        int    `json:"id"`
			Val       string `json:"val"`
			Title     string `json:"title"`
			URL       string `json:"url"`
			NeedShare int    `json:"need_share"`
			M3U8URL   string `json:"m3u8_url"`
		} `json:"m3u8_downlist"`
		NewFtpList []struct {
			ID        int    `json:"id"`
			Title     string `json:"title"`
			URL       string `json:"url"`
			NeedShare int    `json:"need_share"`
		} `json:"new_ftp_list"`
		NewM3U8List []struct {
			ID        int    `json:"id"`
			Title     string `json:"title"`
			URL       string `json:"url"`
			NeedShare int    `json:"need_share"`
		} `json:"new_m3u8_list"`
		Mask              string `json:"mask"`
		CanUrge           int    `json:"can_urge"`
		NarrateVideoID    int    `json:"narrate_video_id"`
		NarrateVideoTitle string `json:"narrate_video_title"`
		NarrateVideoURL   string `json:"narrate_video_url"`
	} `json:"data"`
}

type JianPianSearchApiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ID            int    `json:"id"`
		Title         string `json:"title"`
		Score         string `json:"score"`
		Finished      int    `json:"finished"`
		Shared        int    `json:"shared"`
		IsLook        int    `json:"is_look"`
		Standbytime   int    `json:"standbytime"`
		Definition    int    `json:"definition"`
		EpisodesCount string `json:"episodes_count"`
		Thumbnail     string `json:"thumbnail"`
		Tvimg         string `json:"tvimg"`
		Mask          string `json:"mask"`
		CateID        int    `json:"cate_id"`
	} `json:"data"`
}
