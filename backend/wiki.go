package main

import (
	//"fmt"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	
	"golang.org/x/image/draw"
	_ "image/gif"
)

// DATA STRUCTURES
type Page struct {
  Title string
  Body []byte // byte slice. what is expected by the io lib
  Files []string // Array of file names associated with this page
}

// For the index page to display all available pages
type PageListItem struct {
  Title     string
  HasFiles  bool
}

type IndexPage struct {
  Pages []PageListItem // List of page items with metadata
}

// GLOBAL VARIABLES
var templates = template.Must(template.ParseFiles("edit.html", "view.html", "index.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|upload|delete|delete-file)/([a-zA-Z0-9-]+)$")
var filesDir = "./files" // Directory to store uploaded files
var persistentDir = "/app/persistence" // Directory to store persistent storage
var thumbnailsDir = "/app/persistence/thumbnails" // Directory to store thumbnails

// Mutex map for thread-safe page operations
var pageLocks = make(map[string]*sync.Mutex)
var pageLocksMapMutex sync.Mutex

// getPageLock returns a mutex for a specific page, creating it if necessary
func getPageLock(pageName string) *sync.Mutex {
	pageLocksMapMutex.Lock()
	defer pageLocksMapMutex.Unlock()
	
	if pageLocks[pageName] == nil {
		pageLocks[pageName] = &sync.Mutex{}
	}
	return pageLocks[pageName]
}

// enableCORS adds CORS headers to allow requests from the frontend
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "https://abaj.ai")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// corsMiddleware wraps handlers with CORS support
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// generateMissingThumbnails generates thumbnails for existing media files that don't have one
func generateMissingThumbnails() {
	log.Println("Checking for missing thumbnails...")
	
	mediaDir := filepath.Join(persistentDir, "media")
	files, err := os.ReadDir(mediaDir)
	if err != nil {
		log.Printf("Error reading media directory: %v", err)
		return
	}
	
	generated := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		
		// Only process images
		isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" || ext == ".bmp"
		if !isImage {
			continue
		}
		
		// Check if thumbnail exists
		thumbPath := filepath.Join(thumbnailsDir, filename)
		if _, err := os.Stat(thumbPath); err == nil {
			// Thumbnail exists, skip
			continue
		}
		
		// Generate thumbnail
		sourcePath := filepath.Join(mediaDir, filename)
		if err := generateThumbnail(sourcePath, filename); err != nil {
			log.Printf("Error generating thumbnail for %s: %v", filename, err)
		} else {
			generated++
			log.Printf("Generated thumbnail for %s", filename)
		}
	}
	
	if generated > 0 {
		log.Printf("Generated %d thumbnails", generated)
	} else {
		log.Println("All thumbnails up to date")
	}
}

// generateThumbnail creates a thumbnail for an image file
func generateThumbnail(sourcePath, filename string) error {
	// Open source image
	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	// Decode image
	srcImg, format, err := image.Decode(srcFile)
	if err != nil {
		return err
	}
	
	// Calculate thumbnail dimensions (60x60)
	thumbSize := 60
	srcBounds := srcImg.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()
	
	// Create square thumbnail by cropping to center
	var cropBounds image.Rectangle
	if srcWidth > srcHeight {
		// Landscape: crop width
		offset := (srcWidth - srcHeight) / 2
		cropBounds = image.Rect(offset, 0, offset+srcHeight, srcHeight)
	} else {
		// Portrait: crop height
		offset := (srcHeight - srcWidth) / 2
		cropBounds = image.Rect(0, offset, srcWidth, offset+srcWidth)
	}
	
	// Crop to square
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	var croppedImg image.Image
	if si, ok := srcImg.(subImager); ok {
		croppedImg = si.SubImage(cropBounds)
	} else {
		croppedImg = srcImg
	}
	
	// Create thumbnail image
	thumbImg := image.NewRGBA(image.Rect(0, 0, thumbSize, thumbSize))
	
	// Resize using bilinear interpolation
	draw.BiLinear.Scale(thumbImg, thumbImg.Bounds(), croppedImg, croppedImg.Bounds(), draw.Over, nil)
	
	// Create thumbnail directory if needed
	if err := os.MkdirAll(thumbnailsDir, 0755); err != nil {
		return err
	}
	
	// Save thumbnail
	thumbPath := filepath.Join(thumbnailsDir, filename)
	thumbFile, err := os.Create(thumbPath)
	if err != nil {
		return err
	}
	defer thumbFile.Close()
	
	// Encode based on original format
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(thumbFile, thumbImg, &jpeg.Options{Quality: 75})
	case "png":
		err = png.Encode(thumbFile, thumbImg)
	default:
		// Default to JPEG for other formats
		err = jpeg.Encode(thumbFile, thumbImg, &jpeg.Options{Quality: 75})
	}
	
	return err
}

/*
func handler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:]) //slicing drops the leading /
}
*/

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
  m := validPath.FindStringSubmatch(r.URL.Path)
  if m == nil {
    http.NotFound(w, r)
    return "", errors.New("invalid Page Title")
  }
  return m[2], nil // the title is the second subexpression.
}


func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl+".html",p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

// FUNCTION LITERALS and CLOSURES
/*
The closure returned by makeHandler is a function that takes an http.ResponseWriter and http.Request (in other words, an http.HandlerFunc). The closure extracts the title from the request path, and validates it with the validPath regexp. If the title is invalid, an error will be written to the ResponseWriter using the http.NotFound function. If the title is valid, the enclosed handler function fn will be called with the ResponseWriter, Request, and title as arguments.
*/
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  // called a closure. fn is one of the xxxxHandlers
  return func(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
  }
}


func viewHandler(w http.ResponseWriter,r *http.Request, title string) {
  /* transcended with the closures
  title, err := getTitle(w, r)
  if err != nil {
    return
  }*/
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    return
  }
  /* new and improved version above of the below: error handling!
  title := r.URL.Path[len("/view/"):]
  p, _ := loadPage(title)
  */
  renderTemplate(w, "view", p)
  //fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
  /* same regex error handling as in other handlers.
  title := r.URL.Path[len("/edit/"):]
  p, err := loadPage(title)
  */
  /* closures. preventing code repetition.
  title, err := getTitle(w, r)
  if err != nil {
    return
  }*/
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
  /* Hardcoded html:
  fmt.Fprintf(w, "<h1>Editing %s</h1>"+
    "<form action=\"/save/%s\" method=\"POST\">"+
    "<textarea name=\"body\">%s</textarea><br>"+
    "<input type=\"submit\" value=\"Save\">"+
    "</form>",
    p.Title, p.Title, p.Body)
  */
  /* Code repetition:
  t, _ := template.ParseFiles("edit.html")
  t.Execute(w,p)
  */
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
  // Lock this page for thread-safe operations
  lock := getPageLock(title)
  lock.Lock()
  defer lock.Unlock()
  
  /* closured;
  title, err := getTitle(w, r)
  if err != nil {
    return
  }*/
  //title := r.URL.Path[len("/save/"):]
  body := r.FormValue("body")
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title, Body: []byte(body)}
  } else {
    p.Body = []byte(body)
  }
  err = p.save()
  if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
  }
  
  // Immediately back up the file after saving
  go BackupWikiFiles()
  
  http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// uploadHandler handles file uploads for a specific page
func uploadHandler(w http.ResponseWriter, r *http.Request, title string) {
  // Lock this page for thread-safe operations
  lock := getPageLock(title)
  lock.Lock()
  defer lock.Unlock()
  
  if r.Method != "POST" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  // Ensure files directory exists
  pageDirPath := filepath.Join(filesDir, title)
  if err := os.MkdirAll(pageDirPath, 0755); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  // Parse multipart form, 10 << 20 specifies maximum upload of 10 MB files per file
  r.ParseMultipartForm(10 << 20)
  
  // Get all files from form
  files := r.MultipartForm.File["file"]
  if len(files) == 0 {
    http.Error(w, "No files selected", http.StatusBadRequest)
    return
  }

  // Load the page first
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title, Body: []byte{}, Files: []string{}}
  }

  // Process each file
  uploadedFiles := []string{}
  for _, fileHeader := range files {
    file, err := fileHeader.Open()
    if err != nil {
      http.Error(w, "Error opening file: "+err.Error(), http.StatusInternalServerError)
      return
    }
    defer file.Close()

    // Create file in the server
    filePath := filepath.Join(pageDirPath, fileHeader.Filename)
    dst, err := os.Create(filePath)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    defer dst.Close()

    // Copy file contents
    if _, err := io.Copy(dst, file); err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    // Check if file is already in the list
    found := false
    for _, f := range p.Files {
      if f == fileHeader.Filename {
        found = true
        break
      }
    }
    if !found {
      p.Files = append(p.Files, fileHeader.Filename)
      uploadedFiles = append(uploadedFiles, fileHeader.Filename)
    }
  }

  // Save the updated page
  err = p.save()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  // Immediately back up the files after uploading
  go BackupWikiFiles()

  w.WriteHeader(http.StatusOK)
  if len(uploadedFiles) == 1 {
    w.Write([]byte("File uploaded successfully"))
  } else {
    w.Write([]byte(fmt.Sprintf("%d files uploaded successfully", len(uploadedFiles))))
  }
}

// apiGetPageHandler returns page content as JSON
func apiGetPageHandler(w http.ResponseWriter, r *http.Request) {
  enableCORS(w)
  title := r.URL.Query().Get("title")
  if title == "" {
    http.Error(w, "Missing title parameter", http.StatusBadRequest)
    return
  }

  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }

  w.Header().Set("Content-Type", "application/json")
  w.Write([]byte(`{"title":"` + p.Title + `","body":"` + string(p.Body) + `","files":["` + 
    join(p.Files, `","`) + `"]}`))
}

// Helper function to join strings with a separator
func join(s []string, sep string) string {
  if len(s) == 0 {
    return ""
  }
  result := s[0]
  for _, v := range s[1:] {
    result += sep + v
  }
  return result
}

func (p *Page) save() error {
  filename := filepath.Join(persistentDir, p.Title + ".txt")
  
  // Write page data to persistent directory
  err := os.WriteFile(filename, p.Body, 0600)
  if err != nil {
    return err
  }
  
  // Write files list if there are any
  if len(p.Files) > 0 {
    filesListFilename := filepath.Join(persistentDir, p.Title + ".files.txt")
    filesContent := join(p.Files, "\n")
    err = os.WriteFile(filesListFilename, []byte(filesContent), 0600)
    if err != nil {
      return err
    }
  }
  
  return nil
}

func loadPage(title string) (*Page, error) {
  filename := filepath.Join(persistentDir, title + ".txt")
  body, err := os.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  
  // Load files list if it exists
  filesListFilename := filepath.Join(persistentDir, title + ".files.txt")
  var files []string
  filesContent, err := os.ReadFile(filesListFilename)
  if err == nil && len(filesContent) > 0 {
    files = regexp.MustCompile(`\r?\n`).Split(string(filesContent), -1)
  }
  
  return &Page{Title: title, Body: body, Files: files}, nil
}

func getAllPages() []PageListItem {
  // Get all .txt files (wiki pages) from persistent directory
  files, err := filepath.Glob(filepath.Join(persistentDir, "*.txt"))
  if err != nil {
    return []PageListItem{}
  }
  
  // Create a slice to store file info for sorting
  type fileInfo struct {
    name    string
    modTime time.Time
  }
  
  fileInfos := make([]fileInfo, 0, len(files))
  
  // Extract titles and get modification time for each file
  for _, file := range files {
    // Skip .files.txt metafiles
    if !strings.HasSuffix(file, ".files.txt") {
      info, err := os.Stat(file)
      if err != nil {
        continue
      }
      
      title := strings.TrimSuffix(filepath.Base(file), ".txt")
      fileInfos = append(fileInfos, fileInfo{
        name:    title,
        modTime: info.ModTime(),
      })
    }
  }
  
  // Sort files by modification time (newest first)
  sort.Slice(fileInfos, func(i, j int) bool {
    return fileInfos[i].modTime.After(fileInfos[j].modTime)
  })
  
  // Create page list items with file attachment info
  pageItems := make([]PageListItem, 0, len(fileInfos))
  for _, info := range fileInfos {
    // Check if this page has attachments
    filesListPath := filepath.Join(persistentDir, info.name + ".files.txt")
    _, err := os.Stat(filesListPath)
    hasFiles := err == nil
    
    pageItems = append(pageItems, PageListItem{
      Title:    info.name,
      HasFiles: hasFiles,
    })
  }
  
  return pageItems
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path != "/" {
    http.NotFound(w, r)
    return
  }
  
  // Handle curl file uploads to root domain -> send to "curl"
  if r.Method == "POST" {
    // Parse multipart form
    err := r.ParseMultipartForm(32 << 20) // 32MB max
    if err != nil {
      http.Error(w, "Error parsing form", http.StatusBadRequest)
      return
    }
    
    if r.MultipartForm != nil && len(r.MultipartForm.File) > 0 {
      // Upload files to "curl" page
      title := "curl"
      
      // Create the page directory if it doesn't exist
      pageDirPath := filepath.Join(filesDir, title)
      if err := os.MkdirAll(pageDirPath, 0755); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      
      // Also create in persistent storage
      persistentPageDir := filepath.Join(persistentDir, "files", title)
      if err := os.MkdirAll(persistentPageDir, 0755); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
      }
      
      var uploadedFiles []string
      
      // Process all file fields
      for _, fileHeaders := range r.MultipartForm.File {
        for _, fileHeader := range fileHeaders {
          // Open uploaded file
          file, err := fileHeader.Open()
          if err != nil {
            log.Printf("Error opening file: %v", err)
            continue
          }
          defer file.Close()
          
          // Create destination file
          destPath := filepath.Join(pageDirPath, fileHeader.Filename)
          dest, err := os.Create(destPath)
          if err != nil {
            log.Printf("Error creating file: %v", err)
            continue
          }
          defer dest.Close()
          
          // Copy file
          if _, err := io.Copy(dest, file); err != nil {
            log.Printf("Error saving file: %v", err)
            continue
          }
          
          uploadedFiles = append(uploadedFiles, fileHeader.Filename)
          log.Printf("Uploaded %s to curl", fileHeader.Filename)
        }
      }
      
      if len(uploadedFiles) > 0 {
        // Load existing page or create new one
        p, err := loadPage(title)
        if err != nil {
          p = &Page{Title: title, Body: []byte(""), Files: []string{}}
        }
        
        // Add new files to existing files list
        existingFiles := make(map[string]bool)
        for _, f := range p.Files {
          existingFiles[f] = true
        }
        
        for _, f := range uploadedFiles {
          if !existingFiles[f] {
            p.Files = append(p.Files, f)
          }
        }
        
        // Save page with updated file list
        if err := p.save(); err != nil {
          log.Printf("Error saving page: %v", err)
        }
        
        // Trigger backup
        go BackupWikiFiles()
        
        w.WriteHeader(http.StatusOK)
        if len(uploadedFiles) == 1 {
          fmt.Fprintf(w, "File uploaded to curl: %s\n", uploadedFiles[0])
        } else {
          fmt.Fprintf(w, "%d files uploaded to curl\n", len(uploadedFiles))
        }
        return
      }
    }
  }
  
  // Default GET behavior - show index
  pages := getAllPages()
  indexPage := &IndexPage{Pages: pages}
  
  err := templates.ExecuteTemplate(w, "index.html", indexPage)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

// deleteHandler handles the deletion of a wiki page
func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Lock this page for thread-safe operations
	lock := getPageLock(title)
	lock.Lock()
	defer lock.Unlock()
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delete the main text file
	filename := title + ".txt"
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Error deleting page: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete files list if it exists
	filesListFilename := title + ".files.txt"
	os.Remove(filesListFilename) // Ignore errors as the file might not exist

	// Delete the files directory if it exists
	pageDirPath := filepath.Join(filesDir, title)
	if _, err := os.Stat(pageDirPath); err == nil {
		if err := os.RemoveAll(pageDirPath); err != nil {
			log.Printf("Error removing files directory for %s: %v", title, err)
		}
	}

	// Also remove from persistence if possible
	persistentPath := filepath.Join(persistentDir, filename)
	os.Remove(persistentPath) // Ignore errors
	
	persistentFilesList := filepath.Join(persistentDir, filesListFilename)
	os.Remove(persistentFilesList) // Ignore errors
	
	persistentFilesDir := filepath.Join(persistentDir, "files", title)
	os.RemoveAll(persistentFilesDir) // Ignore errors

	http.Redirect(w, r, "/", http.StatusFound)
}

// deleteFileHandler handles the deletion of a specific file attachment
func deleteFileHandler(w http.ResponseWriter, r *http.Request, title string) {
	// Lock this page for thread-safe operations
	lock := getPageLock(title)
	lock.Lock()
	defer lock.Unlock()
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileName := r.FormValue("filename")
	if fileName == "" {
		http.Error(w, "Missing filename parameter", http.StatusBadRequest)
		return
	}

	// First, remove the file from the filesystem
	filePath := filepath.Join(filesDir, title, fileName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Error deleting file: %v", err), http.StatusInternalServerError)
		return
	}

	// Then, update the page's files list
	p, err := loadPage(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove the file from the Files slice
	var updatedFiles []string
	for _, f := range p.Files {
		if f != fileName {
			updatedFiles = append(updatedFiles, f)
		}
	}
	p.Files = updatedFiles

	// Save the updated page
	if err := p.save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Immediately back up the files after deletion
	go BackupWikiFiles()

	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// mediaUploadHandler handles media uploads (photos and videos)
func mediaUploadHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse multipart form (100MB max memory for videos)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	
	mediaDir := filepath.Join(persistentDir, "media")
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files provided", http.StatusBadRequest)
		return
	}
	
	uploadedCount := 0
	for _, fileHeader := range files {
		// Validate file type (images and videos)
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		validExtensions := map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true,
			".gif": true, ".webp": true, ".bmp": true,
			".mp4": true, ".webm": true, ".mov": true,
			".avi": true, ".mkv": true, ".m4v": true,
		}
		
		if !validExtensions[ext] {
			continue // Skip unsupported files
		}
		
		// Open uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening uploaded file: %v", err)
			continue
		}
		defer file.Close()
		
		// Create destination file
		destPath := filepath.Join(mediaDir, fileHeader.Filename)
		dest, err := os.Create(destPath)
		if err != nil {
			log.Printf("Error creating destination file: %v", err)
			continue
		}
		defer dest.Close()
		
		// Copy file
		if _, err := io.Copy(dest, file); err != nil {
			log.Printf("Error saving file: %v", err)
			continue
		}
		
		// Generate thumbnail for images
		isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" || ext == ".bmp"
		if isImage {
			if err := generateThumbnail(destPath, fileHeader.Filename); err != nil {
				log.Printf("Warning: Could not generate thumbnail for %s: %v", fileHeader.Filename, err)
				// Continue anyway - thumbnail generation is not critical
			} else {
				log.Printf("Generated thumbnail for %s", fileHeader.Filename)
			}
		}
		
		uploadedCount++
		log.Printf("Uploaded media: %s", fileHeader.Filename)
	}
	
	if uploadedCount == 0 {
		http.Error(w, "No valid media files uploaded", http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully uploaded %d file(s)", uploadedCount)
}

// shareTargetHandler handles files shared from mobile devices via Share Target API
func shareTargetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse multipart form
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	
	title := "shared"
	
	// Create the page directory if it doesn't exist
	pageDirPath := filepath.Join(filesDir, title)
	if err := os.MkdirAll(pageDirPath, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Also create in persistent storage
	persistentPageDir := filepath.Join(persistentDir, "files", title)
	if err := os.MkdirAll(persistentPageDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	var uploadedFiles []string
	sharedText := r.FormValue("text")
	sharedURL := r.FormValue("url")
	sharedTitle := r.FormValue("title")
	
	// Process shared files
	if r.MultipartForm != nil && len(r.MultipartForm.File["files"]) > 0 {
		for _, fileHeader := range r.MultipartForm.File["files"] {
			// Open uploaded file
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Error opening shared file: %v", err)
				continue
			}
			defer file.Close()
			
			// Create destination file
			destPath := filepath.Join(pageDirPath, fileHeader.Filename)
			dest, err := os.Create(destPath)
			if err != nil {
				log.Printf("Error creating file: %v", err)
				continue
			}
			defer dest.Close()
			
			// Copy file
			if _, err := io.Copy(dest, file); err != nil {
				log.Printf("Error saving file: %v", err)
				continue
			}
			
			uploadedFiles = append(uploadedFiles, fileHeader.Filename)
			log.Printf("Shared file uploaded: %s", fileHeader.Filename)
		}
	}
	
	// Load or create the "shared" page
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title, Body: []byte(""), Files: []string{}}
	}
	
	// Add shared text/URL to page body if provided, or timestamp if just files
	var newContent string
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	if sharedTitle != "" || sharedText != "" || sharedURL != "" {
		newContent = fmt.Sprintf("\n---\nShared at %s\n", timestamp)
		if sharedTitle != "" {
			newContent += fmt.Sprintf("Title: %s\n", sharedTitle)
		}
		if sharedText != "" {
			newContent += fmt.Sprintf("Text: %s\n", sharedText)
		}
		if sharedURL != "" {
			newContent += fmt.Sprintf("URL: %s\n", sharedURL)
		}
		p.Body = append(p.Body, []byte(newContent)...)
	} else if len(uploadedFiles) > 0 {
		// If only files shared (no text), add a simple timestamp to update mod time
		newContent = fmt.Sprintf("\n---\nFiles shared at %s\n", timestamp)
		p.Body = append(p.Body, []byte(newContent)...)
	}
	
	// Add uploaded files to the page
	existingFiles := make(map[string]bool)
	for _, f := range p.Files {
		existingFiles[f] = true
	}
	
	for _, f := range uploadedFiles {
		if !existingFiles[f] {
			p.Files = append(p.Files, f)
		}
	}
	
	// Save the page
	if err := p.save(); err != nil {
		log.Printf("Error saving shared page: %v", err)
		http.Error(w, "Error saving", http.StatusInternalServerError)
		return
	}
	
	// Trigger backup
	go BackupWikiFiles()
	
	// Redirect to the shared page
	http.Redirect(w, r, "/view/shared", http.StatusSeeOther)
}

// mediaPageHandler serves the media.html page
func mediaPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./media.html")
}

// mediaListHandler returns a JSON list of all media files in the media directory
func mediaListHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	mediaDir := filepath.Join(persistentDir, "media")
	
	// Create media directory if it doesn't exist
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	files, err := os.ReadDir(mediaDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	var mediaFiles []map[string]string
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			ext := strings.ToLower(filepath.Ext(name))
			mediaType := "image"
			
			// Check if it's a video
			if ext == ".mp4" || ext == ".webm" || ext == ".mov" || ext == ".avi" || ext == ".mkv" || ext == ".m4v" {
				mediaType = "video"
			}
			
			mediaFiles = append(mediaFiles, map[string]string{
				"name": name,
				"type": mediaType,
			})
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"media": mediaFiles,
	})
}

// mediaDeleteHandler handles deletion of media (requires authentication)
func mediaDeleteHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse JSON request body
	type DeleteRequest struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		Files    []string `json:"files"`
	}
	
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Authenticate
	if req.Username != "aj" || req.Password != "red" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Delete files
	mediaDir := filepath.Join(persistentDir, "media")
	for _, filename := range req.Files {
		// Security: prevent path traversal
		if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}
		
		filePath := filepath.Join(mediaDir, filename)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Error deleting %s: %v", filename, err)
		}
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Files deleted successfully"))
}

func main() {
  // Create files directory if it doesn't exist
  if err := os.MkdirAll(filesDir, 0755); err != nil {
    log.Fatal(err)
  }
  
  // Create media directory in persistent storage if it doesn't exist
  mediaDir := filepath.Join(persistentDir, "media")
  if err := os.MkdirAll(mediaDir, 0755); err != nil {
    log.Fatal(err)
  }
  
  // Create thumbnails directory
  if err := os.MkdirAll(thumbnailsDir, 0755); err != nil {
    log.Fatal(err)
  }
  
  // Generate thumbnails for existing media files
  go generateMissingThumbnails()

  // Set up file watcher to periodically backup wiki files
  SetupFileWatcher()

  // Set up static file server for uploaded files
  fileServer := http.FileServer(http.Dir(filesDir))
  http.Handle("/files/", http.StripPrefix("/files/", corsMiddleware(fileServer)))

  // Set up static file server for icon files
  iconServer := http.FileServer(http.Dir("./icon"))
  http.Handle("/icon/", http.StripPrefix("/icon/", corsMiddleware(iconServer)))
  
  // Set up static file server for media (from persistent storage)
  mediaFileServer := http.FileServer(http.Dir(mediaDir))
  http.Handle("/media/", http.StripPrefix("/media/", corsMiddleware(mediaFileServer)))
  
  // Set up static file server for thumbnails
  thumbnailsFileServer := http.FileServer(http.Dir(thumbnailsDir))
  http.Handle("/thumbnails/", http.StripPrefix("/thumbnails/", corsMiddleware(thumbnailsFileServer)))
  
  // Serve favicon.ico directly from the icon directory
  http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./icon/favicon.ico")
  })
  
  // Serve manifest.json for PWA and Share Target API
  http.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    http.ServeFile(w, r, "./manifest.json")
  })

  // Root handler
  http.HandleFunc("/", rootHandler)

  // API endpoints
  http.HandleFunc("/api/page", apiGetPageHandler)
  http.HandleFunc("/api/media/list", mediaListHandler)
  http.HandleFunc("/api/media/upload", mediaUploadHandler)
  http.HandleFunc("/api/media/delete", mediaDeleteHandler)
  http.HandleFunc("/api/share", shareTargetHandler)

  // Media page
  http.HandleFunc("/media", mediaPageHandler)

  // Traditional wiki endpoints
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  http.HandleFunc("/upload/", makeHandler(uploadHandler))
  http.HandleFunc("/delete/", makeHandler(deleteHandler))
  http.HandleFunc("/delete-file/", makeHandler(deleteFileHandler))
  
  log.Println("Starting server on http://localhost:21313")
  log.Fatal(http.ListenAndServe(":21313", nil))
}
