package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func GetOpenAPISpec(c *gin.Context) {
	// Try multiple paths to be robust against CWD (backend/ vs root)
	paths := []string{
		filepath.Join("..", "specs", "001-pdf-floorplan-conversion", "contracts", "api.yaml"), // Run from backend/
		filepath.Join("specs", "001-pdf-floorplan-conversion", "contracts", "api.yaml"),     // Run from root
	}

	var content []byte
	var err error

	for _, p := range paths {
		content, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read openapi spec", "details": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/x-yaml", content)
}

func GetSwaggerUI(c *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <meta name="description" content="SwaggerUI" />
  <title>SwaggerUI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '/api/v1/openapi.yaml',
      dom_id: '#swagger-ui',
    });
  };
</script>
</body>
</html>
`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
