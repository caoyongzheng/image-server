package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/caoyongzheng/gotest/utils/fileutil"
	"github.com/caoyongzheng/image-server/utils"
	"github.com/nfnt/resize"
)

// ImgDir 图片存放文件夹
var ImgDir string

// ContentType 图片类型
var ContentType = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"gif":  "image/gif",
	"png":  "image/png",
}
var p = flag.Bool("p", false, "product envoriment")

func main() {
	flag.Parse()
	if *p {
		ImgDir = "/resource/images"
	} else {
		ImgDir = "/Users/caoyongzheng/Pictures/image-server"
	}
	log.Println(ImgDir)
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
	// parse file
	img, imgH, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer img.Close()

	suffix := "jpg"
	if imgH.Filename != "" {
		suffix = strings.ToLower(filepath.Ext(imgH.Filename)[1:])
	}
	if _, ok := ContentType[suffix]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("content type is not support"))
		return
	}

	//获取上传图片哈希值
	imgData, err := ioutil.ReadAll(img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	md5Inst := md5.New()
	md5Inst.Write(imgData)
	imgHash := hex.EncodeToString(md5Inst.Sum(nil))
	filename := imgHash + "." + suffix

	absPath := path.Join(ImgDir, utils.GetImageRePath(filename))

	//判断图片是否已存在
	dir, _ := filepath.Split(absPath)
	os.MkdirAll(dir, os.ModePerm)
	if fileutil.IsExist(absPath) {
		result, _ := json.Marshal(map[string]interface{}{"success": true, "name": filename})
		w.Write(result)
		return
	}

	// create File
	imgFile, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		result, _ := json.Marshal(map[string]interface{}{"success": false, "error": err.Error()})
		w.Write(result)
		return
	}
	defer imgFile.Close()
	imgFile.Write(imgData)
	result, _ := json.Marshal(map[string]interface{}{"success": true, "name": filename})
	w.Write(result)
}

func getImage(w http.ResponseWriter, r *http.Request) {
	var err error
	query := r.URL.Query()
	scale := query.Get("s")
	quality := query.Get("q")
	n := query.Get("n")
	if n == "" || len(n) != 36 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("n is not right"))
		return
	}

	absPath := filepath.Join(ImgDir, utils.GetImageRePath(n))

	// checked file exist
	if !utils.FileExist(absPath) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("image not exist"))
		return
	}

	// check Content type
	ct, ok := ContentType[strings.Split(n, ".")[1]] // t contentType
	if !ok {
		ct = ContentType["jpg"]
	}

	var img image.Image
	var q int
	var ss float64

	// check quality
	if quality != "" && ct == ContentType["jpg"] {
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

	img, err = loadImage(absPath)
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
