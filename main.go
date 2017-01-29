package main

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nfnt/resize"
)

// ImgDir 图片存放文件夹
const ImgDir = "/Users/caoyongzheng/Pictures/image-server"

// ContentType 图片类型
var ContentType = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"gif":  "image/gif",
	"png":  "image/png",
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getImage(w, r)
		} else if r.Method == http.MethodPost {
			addImage(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
	http.ListenAndServe("0.0.0.0:8004", nil)
}

func addImage(w http.ResponseWriter, r *http.Request) {

}

func getImage(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Join(ImgDir, r.URL.Path)

	// checked file exist
	s, err := os.Stat(filename)
	if (err != nil && !os.IsExist(err)) || s.IsDir() {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// check Content type
	ct, ok := ContentType[filepath.Ext(filename)[1:]]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("contentType not supported"))
		return
	}

	var img image.Image
	var q int
	var ss float64

	query := r.URL.Query()
	scale := query.Get("s")
	quality := query.Get("q")

	// check quality
	if quality != "" && ct == "image/jpeg" {
		q, err = strconv.Atoi(quality)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("faild to parse quality"))
			return
		}
	} else {
		q = jpeg.DefaultQuality
	}

	// check scale
	if scale != "" {
		ss, err = strconv.ParseFloat(scale, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("faild to parse scale"))
			return
		}
	}

	img, err = loadImage(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if ss > 0 {
		// 缩略图的大小
		bound := img.Bounds()
		width := uint(float64(bound.Dx()) * ss)
		height := uint(float64(bound.Dy()) * ss)
		img = resize.Resize(width, height, img, resize.Lanczos3)
	}

	var header = w.Header()
	header.Set("Content-Type", ct)
	header.Set("Cache-Control", "max-age=86400")
	switch ct {
	case "image/jpeg":
		o := &jpeg.Options{Quality: q}
		jpeg.Encode(w, img, o)
	case "image/png":
		png.Encode(w, img)
	case "image/gif":
		gif.Encode(w, img, nil)
	}
}

func loadImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	img, _, err = image.Decode(file)
	return
}
