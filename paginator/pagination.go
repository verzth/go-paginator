package paginator

import (
	"fmt"
	"gorm.io/gorm"
	"math"
	"net/http"
	"reflect"
)

type Param struct {
	DB      *gorm.DB
	Req		*http.Request
	Page    int
	Limit   int
	OrderBy []string
	ShowSQL bool
}

type Pagination struct {
	From int `json:"from"`
	To int `json:"to"`
	Total int64 `json:"total"`
	Data interface{} `json:"data"`
	PerPage int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	Offset int `json:"-"`
	FirstPageUrl string `json:"first_page_url"`
	PrevPage *int `json:"prev_page"`
	PrevPageUrl *string `json:"prev_page_url"`
	NextPage *int `json:"next_page"`
	NextPageUrl *string `json:"next_page_url"`
	LastPage int `json:"last_page"`
	LastPageUrl string `json:"last_page_url"`
	Path string `json:"path"`
}

func Paginate(p *Param, result interface{}) *Pagination {
	db := p.DB

	if p.ShowSQL {
		db = db.Debug()
	}
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 25
	}
	if len(p.OrderBy) > 0 {
		for _, o := range p.OrderBy {
			db = db.Order(o)
		}
	}

	done := make(chan bool, 1)
	var paginate Pagination
	var countInPage int
	var count int64
	var offset int

	go countRecords(db, result, done, &count)

	if p.Page == 1 {
		offset = 0
	} else {
		offset = (p.Page - 1) * p.Limit
	}
	db.Limit(p.Limit).Offset(offset).Find(result)

	indirect := reflect.ValueOf(result)
	if indirect.IsValid() && indirect.Elem().Kind() == reflect.Slice {
		countInPage = indirect.Elem().Len()
	}
	<-done

	paginate.FirstPageUrl = fmt.Sprintf("%s%s?page=%d", p.Req.Host, p.Req.URL.Path, 1)
	paginate.Path = fmt.Sprintf("%s%s", p.Req.Host, p.Req.URL.Path)

	paginate.Total = count
	paginate.Data = result
	paginate.CurrentPage = p.Page

	paginate.Offset = offset
	paginate.PerPage = p.Limit
	paginate.LastPage = int(math.Ceil(float64(count) / float64(p.Limit)))
	paginate.LastPageUrl = fmt.Sprintf("%s%s?page=%d", p.Req.Host, p.Req.URL.Path, paginate.LastPage)
	if countInPage > 0 {
		paginate.From = offset+1
		paginate.To = offset+countInPage
	} else {
		paginate.From = 0
		paginate.To = 0
	}

	if p.Page > 1 {
		prevPage := p.Page - 1
		prevPageUrl := fmt.Sprintf("%s%s?page=%d", p.Req.Host, p.Req.URL.Path, prevPage)
		paginate.PrevPage = &prevPage
		paginate.PrevPageUrl = &prevPageUrl
	}

	if p.Page < paginate.LastPage {
		nextPage := p.Page + 1
		nextPageUrl := fmt.Sprintf("%s%s?page=%d", p.Req.Host, p.Req.URL.Path, nextPage)
		paginate.NextPage = &nextPage
		paginate.NextPageUrl = &nextPageUrl
	}
	return &paginate
}

func countRecords(db *gorm.DB, anyType interface{}, done chan bool, count *int64) {
	db.Model(anyType).Count(count)
	done <- true
}