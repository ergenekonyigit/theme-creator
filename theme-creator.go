package emacsthemecreator

import (
	"bytes"
	"encoding/json"
	"fmt"
	htemplate "html/template"
	"io"
	"net/http"
	"strconv"
	"text/template"

	"github.com/lucasb-eyer/go-colorful"
)

const test = `'((mainbg . {{.bg1}}) (mainfg . {{ .fg1}}))`

// RGB struct, contains an uint8 each for
// red, green and blue.
type RGB struct {
	r, g, b int
}

type Theme struct {
	deffacefg        string
	deffacebg        string
	keywordface      string
	builtinface      string
	constantface     string
	commentface      string
	functionnameface string
	stringface       string
	typeface         string
	variableface     string
	warningface      string
}

// NewRGB - Constructor for RGB struct.
// takes 3 uint for red, green and blue and
// returns a new RGB struct
func NewRGB(r, g, b int) RGB {
	rgb := RGB{r: r, g: g, b: b}
	return rgb
}

// rgbToHex - convert 3 rgb color values
// to a hex string.
func rgbToHex(r, g, b int64) (hex string) {
	red := strconv.FormatInt(r, 16)
	green := strconv.FormatInt(g, 16)
	blue := strconv.FormatInt(b, 16)
	colors := []string{red, green, blue}
	for i, j := range colors {
		if len(j) == 1 {
			colors[i] = "0" + j
		}
	}
	return "#" + colors[0] + colors[1] + colors[2]
}

// Lighten - given an rgb value r, multiply each
// value of r by a given factor.
// Example: RGB{122,122,122}.Lighten(0.20) -> RGB{148,148,148}
func (r RGB) Lighten(factor float64) RGB {
	newrgbslice := []int{r.r, r.g, r.b}
	for i, j := range newrgbslice {
		lighter := j + int((factor * (255.0 - float64(j))))
		if lighter > 255 {
			lighter = 255
		}

		newrgbslice[i] = lighter
	}
	return NewRGB(newrgbslice[0], newrgbslice[1], newrgbslice[2])
}

// Darken - given an RGB value r, multiply each
// value of r by factor f, and return a new RGB
// with darker color.
func (r RGB) Darken(factor float64) RGB {
	newrgbslice := []int{r.r, r.g, r.b}
	for i, j := range newrgbslice {
		darker := int(factor * float64(j))
		newrgbslice[i] = darker
	}
	return NewRGB(newrgbslice[0], newrgbslice[1], newrgbslice[2])
}

func decodeTheme(r io.ReadCloser) (*Theme, error) {
	defer r.Close()
	var theme Theme
	err := json.NewDecoder(r).Decode(&theme)
	return &theme, err
}

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/savetheme", saveThemeHandler)
}
func selectedColors(r *http.Request) map[string]string {
	facemap := make(map[string]string)
	faces := []string{"deffacebg", "deffacefg", "builtinface", "stringface",
		"keywordface", "functionnameface", "commentface", "variableface",
		"typeface", "constantface", "warningface"}
	for _, i := range faces {
		facemap[i] = r.FormValue(i)
	}
	facemap = addColors(facemap)
	for key, value := range facemap {
		facemap[key] = "\"" + value + "\""
	}
	facemap["themename"] = r.FormValue("themename")
	return facemap
}
func addColors(colors map[string]string) map[string]string {
	fg := colors["deffacefg"]
	bg := colors["deffacebg"]
	fgcol, _ := colorful.Hex(fg)
	bgcol, _ := colorful.Hex(bg)
	bl, _ := colorful.Hex("#000000")
	darkerfg1 := fgcol.BlendLab(bl, 0.1)
	bg2 := bgcol.BlendLab(bl, 0.05).Hex()
	bg3 := bgcol.BlendLab(bl, 0.1).Hex()
	hex := darkerfg1.Hex()
	colors["fore2"] = hex
	colors["back2"] = bg2
	colors["back3"] = bg3
	return colors
}
func hasDarkBg(c *colorful.Color) bool {
	l, _, _ := c.Lab()
	return l < 0.5
}

func saveThemeHandler(w http.ResponseWriter, r *http.Request) {

	fmap := selectedColors(r)
	var res bytes.Buffer
	themefile := template.Must(template.New("theme").ParseFiles("static/themetemplate.txt"))
	if err := themefile.ExecuteTemplate(&res, "theme", fmap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	str := res.String()
	webview := htemplate.Must(htemplate.New("_displaytheme").ParseFiles("static/displaytheme.html"))
	if err := webview.ExecuteTemplate(w, "_displaytheme", map[string]interface{}{"emacstheme": str}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func handleTheme(r *http.Request) (interface{}, error) {
	switch r.Method {
	case "POST":
		theme, err := decodeTheme(r.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println(theme)

	case "GET":
		var theme Theme
		return theme, nil
	}

	return nil, fmt.Errorf("method not implemented")
}
func handler(w http.ResponseWriter, r *http.Request) {
	val, err := handleTheme(r)
	if err == nil {
		fmt.Fprintf(w, "Success")
		fmt.Println(val)
	}

}

// func main() {
// 	temp := template.Must(template.New("test").Parse(test))
// 	temp.Execute(os.Stdout, map[string]interface{}{"bg1": "\"#000000\"",
// 		"fg1": "\"#ffffff\""})
// 	fmt.Println()
// 	fmt.Println(rgbToHex(255, 255, 255))
// 	fmt.Println(rgbToHex(0, 0, 0))
// 	fmt.Println(122 * 120 / 100)
// }
