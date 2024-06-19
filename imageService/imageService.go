package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pewe21/library"
	"golang.org/x/image/draw"
)

type ImageService struct {
	host         string
	port         string
	internalPort string
}

func NewImageService(cfg AppConfig) *ImageService {
	return &ImageService{
		host:         cfg.Host,
		port:         cfg.Port,
		internalPort: cfg.InternalPort,
	}
}

func (s *ImageService) RegisterRoutes(r *mux.Router) {
	// v1/image
	r.HandleFunc("/", library.CreateHandler(library.JWTMiddleware(s.handleCreateImage))).Methods(http.MethodPost)

	// v1/image/original/{filename}
	r.HandleFunc("/original/{filename}", library.CreateHandler(s.handleGetOriImage)).Methods(http.MethodGet)

	// v1/image/thumbnail/{filename}
	r.HandleFunc("/thumbnail/{filename}", library.CreateHandler(s.handleGetThumbnailImage)).Methods(http.MethodGet)
}

func (s *ImageService) handleGetOriImage(w http.ResponseWriter, r *http.Request) (int, error) {
	vars := mux.Vars(r)
	fileName := vars["filename"]
	file, err := os.Open(filepath.Join("data", "original", fileName))
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer file.Close()
	//set header
	w.Header().Set("Content-Type", "image/jpeg")
	_, err = io.Copy(w, file)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (s *ImageService) handleGetThumbnailImage(w http.ResponseWriter, r *http.Request) (int, error) {
	vars := mux.Vars(r)
	fileName := vars["filename"]
	file, err := os.Open(filepath.Join("data", "thumbnail", fileName))
	if err != nil {
		return http.StatusNotFound, err
	}
	defer file.Close()
	//set header
	w.Header().Set("Content-Type", "image/jpeg")
	_, err = io.Copy(w, file)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (s *ImageService) handleCreateImage(w http.ResponseWriter, r *http.Request) (int, error) {
	println("hit handleCreateImage")

	//ambil image dari formdata
	if err := r.ParseMultipartForm(2 << 20); err != nil { //2MB
		log.Println("Error when parsing request formdata:", err)
		return http.StatusBadRequest, fmt.Errorf("invalid formdata, or missing the required field")
	}

	//baca dengan io.reader
	file, handler, err := r.FormFile("reqImage")
	if err != nil {
		log.Println("Error when creating file handler:", err)
		return http.StatusBadRequest, fmt.Errorf("invalid/missing image")
	}

	defer file.Close()

	fileExt := strings.ToLower(filepath.Ext(handler.Filename))
	if fileExt != ".jpeg" && fileExt != ".jpg" {
		return http.StatusBadRequest, fmt.Errorf("invalid file type: %s, only jpg/jpeg supported", fileExt)
	}

	originalDir := filepath.Join("data", "original")
	thumbDir := filepath.Join("data", "thumbnail")

	err = os.MkdirAll(originalDir, os.ModePerm)
	if err != nil {
		log.Println("Error when creating original image directory:", err)
		return http.StatusInternalServerError, fmt.Errorf("something went wrong")
	}

	err = os.MkdirAll(thumbDir, os.ModePerm)
	if err != nil {
		log.Println("Error when creating thumbnail image directory:", err)
		return http.StatusInternalServerError, fmt.Errorf("something went wrong")
	}

	timestamp := time.Now().Unix()
	uniqueId := uuid.NewString()
	stamp := fmt.Sprintf("%d-%s%s", timestamp, uniqueId, fileExt)

	log.Println("stamp:", stamp)

	// save original file
	oriFilename := filepath.Join("data", "original", handler.Filename)
	log.Println("oriFilename", oriFilename)
	oriFileData, err := os.Create(oriFilename)
	if err != nil {
		log.Println("Error when creating original file object:", err)
		return http.StatusInternalServerError, fmt.Errorf("something went wrong")
	}

	defer oriFileData.Close()

	// copy thumbnail
	_, err = io.Copy(oriFileData, file)
	if err != nil {
		log.Println("Error when duplicating file object:", err)
		return http.StatusInternalServerError, fmt.Errorf("something went wrong")
	}

	//copy thumbnail
	thumbFilename := filepath.Join("data", "thumbnail", handler.Filename)
	thumbFileData, err := os.Create(thumbFilename)
	if err != nil {
		log.Println("Error when creating thumbnail file object")
		return http.StatusInternalServerError, fmt.Errorf("something went wrong")
	}

	defer thumbFileData.Close()

	oriFileData.Seek(0, 0)
	img, _, err := image.Decode(oriFileData)
	if err != nil {
		// log.Println("extensions", fileExt)
		log.Println("Error when decoding original image file:", err)
		return http.StatusInternalServerError, fmt.Errorf("something went wrong")
	}
	//cek apakah ukuran lebih besar dari thumbnail, jika tidak resize ke ukuran thumbnail
	if img.Bounds().Dx() > 512 {
		thumb := resize(img, 512)
		if err = jpeg.Encode(thumbFileData, thumb, nil); err != nil {
			log.Println("Error when scaling image:", err)
			return http.StatusInternalServerError, fmt.Errorf("Invalid image/image is not supported")
		}
	} // jika error masuk ke else

	_, err = oriFileData.Seek(0, 0)
	if err != nil {
		log.Println("Error when seeking original image file:", err)
		return http.StatusInternalServerError, fmt.Errorf("something went wrong")
	}

	//save thumbnail
	_, err = io.Copy(thumbFileData, oriFileData)
	if err != nil {
		log.Println("Error when duplicating thumbnail file object:", err)
		return http.StatusInternalServerError, fmt.Errorf("something went wrong")
	}

	data := AppImage{
		Filename: handler.Filename,
	}

	resp := library.NewResp("Image created", data)

	if err = library.WriteJson(w, http.StatusCreated, resp); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func (s *ImageService) getImage(w http.ResponseWriter, r *http.Request, imageType string) (int, error) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	log.Println(filename)

	// open filename
	originalFilePath := filepath.Join("data", imageType, filename)
	oriFile, err := os.Open(originalFilePath)
	if err != nil {
		return http.StatusNotFound, err
	}

	defer oriFile.Close()

	// set header
	w.Header().Set("Content-Type", "image/jpeg")

	// return imagenya
	_, err = io.Copy(w, oriFile)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func resize(img image.Image, width int) image.Image {
	originWidth := img.Bounds().Dx()
	originHeight := img.Bounds().Dy()
	ration := float64(originHeight) / float64(originWidth)
	newHeight := int(math.Floor(float64(width) * ration))
	rect := image.Rect(0, 0, width, newHeight)
	resized := image.NewRGBA(rect)
	draw.NearestNeighbor.Scale(resized, rect, img, img.Bounds(), draw.Over, nil)

	return resized
}
