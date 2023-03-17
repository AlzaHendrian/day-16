package middleware

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

func UploadFile(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Catch uploadImage Value
		file, err := c.FormFile("uploadImage") //name image di add dan edit project
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		// Open the uploaded File
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		defer src.Close()
		// create a temporary file and store the file into following statements
		tempFile, err := ioutil.TempFile("uploads", "*.png") // => uploads/2fafwv424f.png
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		defer tempFile.Close()

		// Writing the uploaded file into a temporary file
		if _, err = io.Copy(tempFile, src); err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		// retrieve the file name only without "uploads/""
		data := tempFile.Name()
		fileName := data[8:] // => 2fafwv424f.png
		c.Set("dataFile", fileName)

		return next(c)
	}
}
