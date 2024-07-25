package main

import (
	"fmt"
	"net/http"
)

func (api *apiConfig) getMetrics(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "text/html")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(fmt.Sprintf(`
    <html>

    <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
    </body>

    </html>`, api.fileserverHits)))
}
