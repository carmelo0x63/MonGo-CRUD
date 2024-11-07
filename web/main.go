package main

import (
    "html/template"
    "log"
    "net/http"
    "os"
//    "path/filepath"
)

func main() {
    r := http.NewServeMux()
    r.HandleFunc("/", handleIndex)
    r.HandleFunc("/static/", handleStatic)

    fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.HandleFunc("/", handleIndex)

    // Get host and port from environment variables with defaults
    host := getEnvOrDefault("WEB_HOST", "0.0.0.0")
    port := getEnvOrDefault("WEB_PORT", "3000")
    addr := host + ":" + port

    log.Printf("Web Server starting on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("templates/index.html"))
    tmpl.Execute(w, nil)
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
    http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
}
