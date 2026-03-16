package api

import "net/http"

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"name":"GeoKrety Stats API","version":"v3","docs":"/docs","openapi":"/openapi.yaml","health":"/health","metrics":"/metrics","root":"/api/v3"}`))
}
