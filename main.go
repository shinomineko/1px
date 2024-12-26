package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"strconv"
)

type PageData struct {
	Color        string
	Transparency int
	ImagePreview template.HTML
	Base64Data   string
}

var templates = template.Must(template.ParseFiles("index.tmpl"))

func generateColorImage(hexColor string, transparency int) (string, string, error) {
	previewImg := image.NewRGBA(image.Rect(0, 0, 400, 240))

	// the checker pattern
	for x := 0; x < 400; x++ {
		for y := 0; y < 240; y++ {
			isEvenSquareX := (x/20)%2 == 0
			isEvenSquareY := (y/20)%2 == 0
			if isEvenSquareX == isEvenSquareY {
				previewImg.Set(x, y, color.RGBA{200, 200, 200, 255})
			} else {
				previewImg.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	var r, g, b uint8
	fmt.Sscanf(hexColor[1:], "%02x%02x%02x", &r, &g, &b)
	overlayColor := color.RGBA{r, g, b, uint8(transparency)}

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))

	for x := 0; x < 400; x++ {
		for y := 0; y < 400; y++ {
			bgColor := previewImg.RGBAAt(x, y)
			alpha := float64(overlayColor.A) / 255.0

			newR := uint8(float64(bgColor.R)*(1-alpha) + float64(overlayColor.R)*alpha)
			newG := uint8(float64(bgColor.G)*(1-alpha) + float64(overlayColor.G)*alpha)
			newB := uint8(float64(bgColor.B)*(1-alpha) + float64(overlayColor.B)*alpha)

			previewImg.Set(x, y, color.RGBA{newR, newG, newB, 255})
		}
	}

	img.Set(0, 0, overlayColor)

	var previewBuf bytes.Buffer
	if err := png.Encode(&previewBuf, previewImg); err != nil {
		return "", "", err
	}
	previewData := fmt.Sprintf("data:image/png;base64,%s",
		base64.StdEncoding.EncodeToString(previewBuf.Bytes()))

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", err
	}
	base64Data := fmt.Sprintf("data:image/png;base64,%s",
		base64.StdEncoding.EncodeToString(buf.Bytes()))

	return previewData, base64Data, nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	color := r.URL.Query().Get("color")
	if color == "" {
		color = "#000000"
	}

	transparency := 255
	if t := r.URL.Query().Get("transparency"); t != "" {
		if val, err := strconv.Atoi(t); err == nil {
			transparency = val
		}
	}

	previewData, base64Data, err := generateColorImage(color, transparency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Color:        color,
		Transparency: transparency,
		ImagePreview: template.HTML(fmt.Sprintf("<img src=\"%s\">", previewData)),
		Base64Data:   base64Data,
	}

	templates.ExecuteTemplate(w, "index.tmpl", data)
}

func main() {
	http.HandleFunc("/", handleRequest)
	fmt.Println("Server starting on port 3939")
	http.ListenAndServe(":3939", nil)
}
