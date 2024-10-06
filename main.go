package main

import (
    "io"
    "net/http"
    "net/url"
    "log"
    "os"
)

func handleProxy(w http.ResponseWriter, r *http.Request) {
    targetURL := r.URL.Query().Get("url")
    if targetURL == "" {
        http.Error(w, "URL parameter is missing", http.StatusBadRequest)
        return
    }

    parsedURL, err := url.Parse(targetURL)
    if err != nil {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }

    req, err := http.NewRequest(r.Method, parsedURL.String(), r.Body)
    if err != nil {
        http.Error(w, "Failed to create request", http.StatusInternalServerError)
        return
    }

    // Copy headers from the original request
    for key, values := range r.Header {
        for _, value := range values {
            req.Header.Add(key, value)
        }
    }

    proxyURL, _ := url.Parse(os.Getenv("PROXY_URL"))
    client := &http.Client{
        Transport: &http.Transport{
            Proxy: http.ProxyURL(proxyURL),
        },
    }
    resp, err := client.Do(req)
    if err != nil {
        http.Error(w, "Failed to make request", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // Copy response headers
    for key, values := range resp.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }

    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}

func main() {
    http.HandleFunc("/", handleProxy)
    log.Println("Starting proxy server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}