package file

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
	"university-backend/foundation/web"
	"university-backend/internal/service/hashing"
)

type Controller struct {
	*web.App
	fileServerBasePath string
}

type Config struct {
	MediaBaseLink string `conf:"default:./media"`
}

type onlyFilesFS struct {
	http.FileSystem
}

func NewController(app *web.App, fileServerBasePath string) *Controller {
	return &Controller{app, fileServerBasePath}
}

func (cf Controller) File(c *gin.Context) {
	fs := gin.Dir("./media", false)
	if _, noListing := fs.(*onlyFilesFS); noListing {
		c.Writer.WriteHeader(http.StatusNotFound)
	}

	file := c.Param("filepath")
	if !strings.Contains(file[1:], "/") {
		OpenH := hashing.OpenHash(file)
		list := strings.Split(OpenH, " ")
		if len(list) == 3 {
			linkTime, err := time.Parse("02.01.2006 15:04:05 ", list[1]+" "+list[2])
			if err != nil {
				c.JSON(http.StatusBadRequest, map[string]any{
					"error":  "incorrect link",
					"status": false,
				})
				return
			}
			if linkTime.Before(time.Now().UTC()) {
				c.JSON(http.StatusBadRequest, map[string]any{
					"error":  "expired link",
					"status": false,
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, map[string]any{
				"error":  "incorrect link",
				"status": false,
			})
			return
		}
		// Check if file exists and/or if we have permission to access it
		f, err := fs.Open(list[0])
		if err != nil {
			c.JSON(http.StatusBadGateway, map[string]any{
				"error":  "file not found" + err.Error(),
				"status": false,
			})
			return
		}
		f.Close()

		http.ServeFile(c.Writer, c.Request, "./media"+list[0])
	} else {
		f, err := fs.Open(file)
		if err != nil {
			c.JSON(http.StatusBadGateway, map[string]any{
				"error":  "file not found",
				"status": false,
			})
			return
		}
		f.Close()

		http.ServeFile(c.Writer, c.Request, "./media"+file)
	}

}
