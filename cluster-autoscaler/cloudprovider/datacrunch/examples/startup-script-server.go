package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	texttmpl "text/template"

	"github.com/google/uuid"
	datacrunchclient "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/datacrunch/datacrunch-go"
)

type ScriptRequest struct {
	// Optionally allow overrides in the future
}

type ScriptResponse struct {
	Script   string `json:"script"`
	ScriptID string `json:"script_id"`
}

func main() {
	http.HandleFunc("/script", handleGenerateScript)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Startup script generator listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleGenerateScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clientID := os.Getenv("DATACRUNCH_CLIENT_ID")
	clientSecret := os.Getenv("DATACRUNCH_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		http.Error(w, "Missing DATACRUNCH_CLIENT_ID or DATACRUNCH_CLIENT_SECRET", http.StatusInternalServerError)
		return
	}
	k3sToken := os.Getenv("K3S_TOKEN")
	k3sUrl := os.Getenv("K3S_URL")
	_ = os.Getenv("INSTALL_K3S_VERSION") // used in expand env

	if k3sToken == "" || k3sUrl == "" {
		http.Error(w, "Missing K3S_TOKEN or K3S_SERVER env vars", http.StatusInternalServerError)
		return
	}

	// Read script template from env or file
	template := os.Getenv("STARTUP_SCRIPT_TEMPLATE")
	if template == "" {
		http.Error(w, "Missing STARTUP_SCRIPT_TEMPLATE env var", http.StatusInternalServerError)
		return
	}

	// Generate a unique script ID
	scriptName := uuid.New().String()

	// Prepare template variables
	vars := map[string]string{
		"K3S_TOKEN":                k3sToken,
		"K3S_URL":                  k3sUrl,
		"SCRIPT_NAME":              scriptName,
		"DATACRUNCH_CLIENT_ID":     clientID,
		"DATACRUNCH_CLIENT_SECRET": clientSecret,
	}

	// Use Go's text/template for variable substitution with default delimiters
	tmpl, err := texttmpl.New("startup").Parse(template)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, vars)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}
	template = sb.String()

	// Upload script to DataCrunch
	client := datacrunchclient.NewClient(clientID, clientSecret)
	scriptID, err := client.UploadStartupScript(scriptName, template)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload script to DataCrunch: %v", err), http.StatusInternalServerError)
		return
	}
	scriptID = strings.TrimSpace(scriptID)

	w.Write([]byte(scriptID))
}
