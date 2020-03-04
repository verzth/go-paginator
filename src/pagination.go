package paginator

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"math"
	"net/http"
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
	Total int `json:"total"`
	Data interface{} `json:"data"`
	PerPage int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	Offset int `json:"-"`
	FirstPageUrl string `json:"first_page_url"`
	PrevPage int `json:"prev_page"`
	PrevPageUrl string `json:"prev_page_url"`
	NextPage int `json:"next_page"`
	NextPageUrl string `json:"next_page_url"`
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
	var count int
	var offset int

	go countRecords(db, result, done, &count)

	if p.Page == 1 {
		offset = 0
	} else {
		offset = (p.Page - 1) * p.Limit
	}

	db.Limit(p.Limit).Offset(offset).Find(result)
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
	if paginate.Total > 0 {
		paginate.From = offset+1
		paginate.To = offset+len(paginate.Data.([]interface{}))
	} else {
		paginate.From = 0
		paginate.To = 0
	}

	if p.Page > 1 {
		paginate.PrevPage = p.Page - 1
	} else {
		paginate.PrevPage = p.Page
	}
	paginate.PrevPageUrl = fmt.Sprintf("%s%s?page=%d", p.Req.Host, p.Req.URL.Path, paginate.PrevPage)

	if p.Page == paginate.LastPage {
		paginate.NextPage = p.Page
	} else {
		paginate.NextPage = p.Page + 1
	}
	paginate.NextPageUrl = fmt.Sprintf("%s%s?page=%d", p.Req.Host, p.Req.URL.Path, paginate.NextPage)
	return &paginate
}

func countRecords(db *gorm.DB, anyType interface{}, done chan bool, count *int) {
	db.Model(anyType).Count(count)
	done <- true
}