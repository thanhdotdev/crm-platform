package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Params holds pagination parameters.
type Params struct {
	Page    int
	PerPage int
}

// Offset returns the database offset based on current page.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// TotalPages calculates total pages from total count.
func (p Params) TotalPages(total int64) int {
	return int(math.Ceil(float64(total) / float64(p.PerPage)))
}

// FromContext extracts pagination params from query string (?page=1&per_page=20).
func FromContext(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(DefaultPerPage)))

	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return Params{Page: page, PerPage: perPage}
}

// Apply applies pagination to a GORM query.
func Apply(db *gorm.DB, params Params) *gorm.DB {
	return db.Offset(params.Offset()).Limit(params.PerPage)
}
