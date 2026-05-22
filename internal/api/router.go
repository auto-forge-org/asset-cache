package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/auto-forge-org/asset-cache/internal/service"
	"github.com/auto-forge-org/asset-cache/internal/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.AssetService
}

func NewHandler(svc *service.AssetService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	r.GET("/healthz", h.health)
	r.GET("/metrics/cache", h.cacheStats)

	v1 := r.Group("/api/v1")
	v1.POST("/assets", h.uploadAsset)
	v1.GET("/assets", h.listAssets)
	v1.GET("/assets/search", h.searchAssets)
	v1.GET("/assets/:id", h.getAsset)
	v1.GET("/assets/:id/download", h.downloadAsset)
	v1.PUT("/assets/:id/version", h.newVersion)
	v1.GET("/assets/:id/versions", h.listVersions)
	v1.POST("/assets/:id/sign", h.signAsset)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) cacheStats(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.CacheStats())
}

func (h *Handler) uploadAsset(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxUploadBytes+1024)

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file exceeds 5GB limit"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileHdr, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'file' field"})
		return
	}
	if fileHdr.Size > service.MaxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file exceeds 5GB limit"})
		return
	}
	f, err := fileHdr.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read upload"})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read upload"})
		return
	}

	var meta map[string]interface{}
	if raw := c.PostForm("metadata"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &meta); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "metadata is not valid JSON"})
			return
		}
	}

	asset, err := h.svc.Upload(service.UploadInput{
		Name:     fileHdr.Filename,
		Type:     fileHdr.Header.Get("Content-Type"),
		UserID:   c.GetHeader("X-User-ID"),
		Data:     data,
		Metadata: meta,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrEmptyFile):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":       asset.ID,
		"url":      asset.StoragePath,
		"checksum": asset.Checksum,
		"asset":    asset,
	})
}

func (h *Handler) listAssets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"assets": h.svc.Search("", nil)})
}

func (h *Handler) searchAssets(c *gin.Context) {
	query := c.Query("query")
	var tags []string
	if raw := c.Query("tags"); raw != "" {
		tags = strings.Split(raw, ",")
	}
	results := h.svc.Search(query, tags)
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

func (h *Handler) getAsset(c *gin.Context) {
	asset, _, err := h.svc.Get(c.Param("id"))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asset)
}

func (h *Handler) downloadAsset(c *gin.Context) {
	id := c.Param("id")
	expStr := c.Query("exp")
	sig := c.Query("sig")
	if expStr == "" || sig == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing signature"})
		return
	}
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exp"})
		return
	}
	if !h.svc.VerifySignature(id, exp, sig) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired signature"})
		return
	}
	asset, data, err := h.svc.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	c.Data(http.StatusOK, asset.Type, data)
}

func (h *Handler) newVersion(c *gin.Context) {
	id := c.Param("id")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxUploadBytes+1024)
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}
	v, err := h.svc.NewVersion(id, data)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		if errors.Is(err, service.ErrEmptyFile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *Handler) listVersions(c *gin.Context) {
	versions, err := h.svc.Versions(c.Param("id"))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"versions": versions, "count": len(versions)})
}

type signRequest struct {
	Expiration     string   `json:"expiration"`
	AllowedDomains []string `json:"allowed_domains"`
}

func (h *Handler) signAsset(c *gin.Context) {
	var req signRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ttl, err := time.ParseDuration(req.Expiration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expiration duration"})
		return
	}
	signed, err := h.svc.Sign(c.Param("id"), ttl)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, signed)
}
