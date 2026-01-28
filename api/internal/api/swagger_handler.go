package api

import (
	"net/http"
	"strings"
)

// HandleSwaggerUI serves the Swagger UI interface
func (s *Server) HandleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	// Serve Swagger UI HTML that loads the OpenAPI spec
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>SprintSpark API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin: 0;
            padding: 0;
        }
        .topbar {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            // Determine the base URL
            const baseUrl = window.location.origin;

            // Initialize Swagger UI
            const ui = SwaggerUIBundle({
                url: baseUrl + "/api/openapi",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                persistAuthorization: true,
                tryItOutEnabled: true,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch', 'head', 'options'],
                validatorUrl: null,
                // Add support for Bearer token and API Key authentication
                onComplete: function() {
                    console.log("Swagger UI loaded successfully");
                }
            });

            window.ui = ui;
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// HandleOpenAPIYAML serves the OpenAPI spec as downloadable YAML
// This endpoint requires authentication (JWT or API Key)
func (s *Server) HandleOpenAPIYAML(w http.ResponseWriter, r *http.Request) {
	// Read OpenAPI spec file
	content, err := s.getOpenAPISpec()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load API specification", "internal_error")
		return
	}

	// Check if user wants to download
	download := r.URL.Query().Get("download")
	if download == "true" || strings.Contains(r.Header.Get("Accept"), "application/octet-stream") {
		w.Header().Set("Content-Disposition", "attachment; filename=\"sprintspark-openapi.yaml\"")
		w.Header().Set("Content-Type", "application/octet-stream")
	} else {
		w.Header().Set("Content-Type", "application/yaml")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// HandleOpenAPIJSON serves the OpenAPI spec as JSON
// This endpoint requires authentication (JWT or API Key)
func (s *Server) HandleOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	// For now, we only have YAML. In the future, we could convert to JSON
	// For simplicity, we'll just serve YAML with a note
	respondError(w, http.StatusNotImplemented, "JSON format not yet supported, use /api/openapi.yaml", "not_implemented")
}
