// Package handlers implements HTTP request handlers for the leaderboard API.
package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/geokrety/leaderboard-api/internal/models"
)

const (
	defaultPageSize = 25
	maxPageSize     = 100
)

// Handler holds shared dependencies for all route handlers.
type Handler struct {
	DB *pgxpool.Pool
}

// New creates a new Handler.
func New(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

// parsePagination extracts page/per_page from query, returning offset and limit.
func parsePagination(c *gin.Context) (page, perPage, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ = strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(defaultPageSize)))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > maxPageSize {
		perPage = defaultPageSize
	}
	offset = (page - 1) * perPage
	return
}

// buildLinks builds pagination links for a collection.
func buildLinks(c *gin.Context, page, perPage int, hasNext bool) *models.Links {
	base := fmt.Sprintf("%s?%s", c.Request.URL.Path, c.Request.URL.RawQuery)
	_ = base
	self := fullURL(c, page, perPage)
	l := &models.Links{Self: self}
	if hasNext {
		l.Next = fullURL(c, page+1, perPage)
	}
	if page > 1 {
		l.Prev = fullURL(c, page-1, perPage)
	}
	return l
}

func fullURL(c *gin.Context, page, perPage int) string {
	q := c.Request.URL.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("per_page", strconv.Itoa(perPage))
	return fmt.Sprintf("%s?%s", c.Request.URL.Path, q.Encode())
}

func totalPages(total int64, perPage int) int {
	return int(math.Ceil(float64(total) / float64(perPage)))
}

// ok sends a successful JSON:API response.
func ok(c *gin.Context, data interface{}, meta models.Meta, links *models.Links) {
	c.JSON(http.StatusOK, models.Response{
		Data:  data,
		Meta:  meta,
		Links: links,
	})
}

// errNotFound sends a 404.
func errNotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"errors": []gin.H{{"status": "404", "title": msg}}})
}

// errInternal sends a 500.
func errInternal(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"errors": []gin.H{{"status": "500", "title": err.Error()}}})
}

// moveTypeName maps move_type int to human-readable name.
func moveTypeName(t int) string {
	switch t {
	case 0:
		return "drop"
	case 1:
		return "grab"
	case 2:
		return "comment"
	case 3:
		return "seen"
	case 4:
		return "archived"
	case 5:
		return "dip"
	default:
		return "unknown"
	}
}

// gkTypeName maps gk_type int to human-readable name.
// Types from GeoKrety PHP constants.
func gkTypeName(t int) string {
	switch t {
	case 0:
		return "Traditional"
	case 1:
		return "Book/CD/DVD"
	case 2:
		return "Human"
	case 3:
		return "Coin"
	case 4:
		return "KretyPost"
	case 5:
		return "Pebble"
	case 6:
		return "Car"
	case 7:
		return "Playing Card"
	case 8:
		return "Dog Tag"
	case 9:
		return "Jigsaw"
	case 10:
		return "Easter Egg"
	default:
		return "Unknown"
	}
}

func avatarRef(bucket, key sql.NullString) *string {
	if !bucket.Valid || !key.Valid || bucket.String == "" || key.String == "" {
		return nil
	}
	value := bucket.String + "/" + key.String
	return &value
}
