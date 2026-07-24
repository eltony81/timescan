package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func alertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Println("==================================================")
	fmt.Println("[ALERT WEBHOOK RECEIVER] Incoming Alert Payload:")
	fmt.Println(string(body))
	fmt.Println("==================================================")

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"received"}`))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/alerts", alertHandler)
	fmt.Printf("[Webhook Receiver] Server listening on :%s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
