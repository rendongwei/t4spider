package common

type PlayResponse struct {
	Header string `json:"header"`
	Parse  string `json:"parse"`
	URL    string `json:"url"`
	Jx     string `json:"jx"`
}

type Vod struct {
	VodPic      string `json:"vod_pic,omitempty"`
	VodContent  string `json:"vod_content,omitempty"`
	TypeName    string `json:"type_name,omitempty"`
	VodPlayFrom string `json:"vod_play_from,omitempty"`
	VodRemarks  string `json:"vod_remarks,omitempty"`
	VodYear     string `json:"vod_year,omitempty"`
	VodID       string `json:"vod_id,omitempty"`
	VodPlayURL  string `json:"vod_play_url,omitempty"`
	VodArea     string `json:"vod_area,omitempty"`
	VodDirector string `json:"vod_director,omitempty"`
	VodName     string `json:"vod_name,omitempty"`
	VodActor    string `json:"vod_actor,omitempty"`
}

type Class struct {
	TypeName string `json:"type_name"`
	TypeID   string `json:"type_id"`
}

type SearchResponse struct {
	List []Vod `json:"list"`
}
type DetailResponse struct {
	List []Vod `json:"list"`
}

type HomeResponse struct {
	List   []Vod       `json:"list"`
	Class  []Class     `json:"class"`
	Filter interface{} `json:"filter"`
}

type CateResponse struct {
	Total     int64 `json:"total"`
	Pagecount int64 `json:"pagecount"`
	Limit     int   `json:"limit"`
	Page      int   `json:"page"`
	List      []Vod `json:"list"`
}

type Spider interface {
	Home() HomeResponse
	Cate(tid string, pg int, ext string) CateResponse
	Detail(id string) DetailResponse
	Play(id string) PlayResponse
	Search(key string, flag bool) SearchResponse
}
