import (
	"github.com/jinzhu/gorm"
)

type Server struct {
	db     *gorm.DB
	router *Router
}