package api

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
)

type rootInfo struct {
	XMLName xml.Name `json:"-" xml:"rootInfo"`
	Name    string   `json:"name" xml:"name"`
	Version string   `json:"version" xml:"version"`
	Docs    string   `json:"docs" xml:"docs"`
	OpenAPI string   `json:"openapi" xml:"openapi"`
	Health  string   `json:"health" xml:"health"`
	Metrics string   `json:"metrics" xml:"metrics"`
	Root    string   `json:"root" xml:"root"`
}

func Home(w http.ResponseWriter, r *http.Request) {
	payload := rootInfo{
		Name:    "GeoKrety Stats API",
		Version: "v3",
		Docs:    "/docs",
		OpenAPI: "/openapi.yaml",
		Health:  "/health",
		Metrics: "/metrics",
		Root:    "/api/v3",
	}
	if acceptsXML(r) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_ = xml.NewEncoder(w).Encode(payload)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func acceptsXML(r *http.Request) bool {
	if r == nil {
		return false
	}
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml")
}
