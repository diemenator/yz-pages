package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

type PathConfig struct {
	projectRoot  string
	targetDir    string
	scriptDir    string
	styleCss     string
	targetIxMd   string
	targetIxJson string
}

func NewPathConfig() PathConfig {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine source file path")
	}

	sourceDir := filepath.Dir(sourceFile)
	absolutePath, err := filepath.Abs(sourceDir)
	if err != nil {
		panic(err)
	}
	scriptDir := absolutePath
	styleCss := filepath.Join(absolutePath, "style.css")
	projectRoot := filepath.Dir(filepath.Dir(absolutePath))

	targetDir := filepath.Join(projectRoot, "meme-gallery")
	targetIxMd := filepath.Join(targetDir, "index.md")
	targetIxJson := filepath.Join(targetDir, "gallery.json")

	return PathConfig{
		projectRoot:  projectRoot,
		targetDir:    targetDir,
		scriptDir:    scriptDir,
		targetIxMd:   targetIxMd,
		targetIxJson: targetIxJson,
		styleCss:     styleCss,
	}
}

// GalleryItem represents a single item in the gallery
type GalleryItem struct {
	Filename string `json:"filename"`
	Caption  string `json:"caption"`
}

// Gallery represents the entire gallery collection
type Gallery struct {
	Items []GalleryItem `json:"items"`
}

// readCSS reads the CSS file and returns its content
func readCSS(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(content)
}

// scanGalleryContent scans the directory for images and builds the gallery model
func scanGalleryContent(galleryDir string) (Gallery, error) {
	var gallery Gallery

	// Get all files in the gallery directory
	files, err := ioutil.ReadDir(galleryDir)
	if err != nil {
		return gallery, err
	}

	// Filter for image files and create gallery items
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// Only include images (jpg, jpeg, png, gif)
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
			// Generate a caption by removing extension and replacing dashes/underscores with spaces
			caption := strings.TrimSuffix(filename, ext)
			caption = strings.ReplaceAll(caption, "-", " ")
			caption = strings.ReplaceAll(caption, "_", " ")
			caption = strings.Title(caption) // Capitalize first letter of each word

			gallery.Items = append(gallery.Items, GalleryItem{
				Filename: filename,
				Caption:  caption,
			})
		}
	}

	return gallery, nil
}

// saveJSON saves the gallery model as JSON
func saveJSON(gallery Gallery, outputPath string) {
	jsonData, err := json.MarshalIndent(gallery, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		panic(err)
	}
}

// generateMarkdown creates a markdown file with the gallery content
func generateMarkdown(gallery Gallery, paths PathConfig) {
	// Read the CSS content
	css := readCSS(paths.styleCss)

	// Create a template for the markdown file
	const markdownTemplate = `# Meme Gallery
A collection of images with unified sizing.
<div class="image-gallery-container"><div class="image-gallery">{{range .Items}}
<div class="gallery-item"><div class="gallery-item-content">
		<div class="image-container"><img src="{{.Filename}}" alt="{{.Caption}}" ></div>
  		<div class="image-caption"><p class="caption">{{.Caption}}</p></div>
</div></div>{{end}}
</div></div></div>
<style>
{{.CSS}}
</style>
`

	// Create a structure to hold both gallery and CSS
	data := struct {
		Items []GalleryItem
		CSS   string
	}{
		Items: gallery.Items,
		CSS:   css,
	}

	// Parse and execute the template
	tmpl, err := template.New("markdown").Parse(markdownTemplate)
	if err != nil {
		panic(err)
	}

	// remove the old file if it exists
	if _, err := os.Stat(paths.targetIxMd); err == nil {
		err = os.Remove(paths.targetIxMd)
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Create(paths.targetIxMd)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		panic(err)
	}
}

func main() {

	// Paths configuration
	paths := NewPathConfig()

	// Scan the gallery directory
	gallery, err := scanGalleryContent(paths.targetDir)
	if err != nil {
		fmt.Printf("Error scanning gallery: %v\n", err)
		os.Exit(1)
	}

	// Save the gallery model as JSON
	saveJSON(gallery, paths.targetIxJson)
	fmt.Printf("Gallery JSON saved to %s\n", paths.targetIxJson)

	// Generate markdown with embedded CSS
	generateMarkdown(gallery, paths)
	fmt.Printf("Gallery markdown saved to %s\n", paths.targetIxMd)
}
