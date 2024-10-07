package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

type URL struct {
    ID         int       `json:"id"`
    Original   string    `json:"original_url"`
    Short      string    `json:"short_url"`
    CreatedAt  time.Time `json:"created_at"`
    ClickCount int       `json:"click_count"`
}

var db *sql.DB

func main() {
    var err error

    // Connect to SQLite database
    db, err = sql.Open("sqlite3", "./urlshortener.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Setup HTTP routes
    http.HandleFunc("/shorten", shortenURLHandler)
    http.HandleFunc("/", redirectHandler)

    fmt.Println("Server started at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// Handler to shorten URLs
func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse the incoming request to get the original URL
    var request struct {
        URL string `json:"url"`
    }
    json.NewDecoder(r.Body).Decode(&request)

    // Generate a short URL
    shortID := generateShortID()

    // Insert into the database
    stmt, err := db.Prepare("INSERT INTO urls(original_url, short_url) VALUES(?, ?)")
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    _, err = stmt.Exec(request.URL, shortID)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // Return the short URL in the response
    shortURL := fmt.Sprintf("http://localhost:8080/%s", shortID)
    json.NewEncoder(w).Encode(map[string]string{"short_url": shortURL})
}

// Handler to redirect to the original URL
func redirectHandler(w http.ResponseWriter, r *http.Request) {
    shortID := r.URL.Path[1:]

    // Query the original URL from the database
    var originalURL string
    err := db.QueryRow("SELECT original_url FROM urls WHERE short_url = ?", shortID).Scan(&originalURL)
    if err != nil {
        http.Error(w, "URL not found", http.StatusNotFound)
        return
    }

    // Increment the click count
    _, err = db.Exec("UPDATE urls SET click_count = click_count + 1 WHERE short_url = ?", shortID)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // Redirect the user to the original URL
    http.Redirect(w, r, originalURL, http.StatusFound)
}

// Helper function to generate a short URL ID (you can improve this)
func generateShortID() string {
    return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000) // Basic random ID generator
}
