package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CaptureResult struct {
	TargetURL  string         `json:"targetUrl"`
	Device     string         `json:"device"`
	FullPage   bool           `json:"fullPage"`
	Dimensions map[string]int `json:"dimensions,omitempty"`
	FileSize   string         `json:"fileSize,omitempty"`
	ImageURL   string         `json:"imageUrl"`
}

type APIResponse struct {
	Status bool           `json:"status"`
	Code   int            `json:"code,omitempty"`
	Error  string         `json:"error,omitempty"`
	Result *CaptureResult `json:"result,omitempty"`
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func normalizeTargetURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		trimmed = "https://" + trimmed
	}
	return trimmed
}

type ProviderResult struct {
	ImageURL string
	Width    int
	Height   int
	FileSize string
	RawBytes []byte
}

func captureWebsite(ctx context.Context, targetURL, device string, fullPage, needRaw bool) (*ProviderResult, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	isMobile := strings.ToLower(device) == "mobile"

	params := url.Values{}
	params.Set("url", targetURL)
	params.Set("screenshot", "true")
	params.Set("meta", "false")
	params.Set("screenshot.fullPage", fmt.Sprintf("%v", fullPage))

	w := 1920
	h := 1080
	if isMobile {
		w = 414
		h = 896
		params.Set("viewport.isMobile", "true")
		params.Set("viewport.width", "414")
		params.Set("viewport.height", "896")
	} else {
		params.Set("viewport.width", "1920")
		params.Set("viewport.height", "1080")
	}

	microURL := "https://api.microlink.io?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, microURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "Essite/1.0")
		res, err := client.Do(req)
		if err == nil && res.StatusCode == http.StatusOK {
			defer res.Body.Close()
			var microData struct {
				Status string `json:"status"`
				Data   struct {
					Screenshot struct {
						URL        string `json:"url"`
						Width      int    `json:"width"`
						Height     int    `json:"height"`
						SizePretty string `json:"size_pretty"`
					} `json:"screenshot"`
				} `json:"data"`
			}
			if err := json.NewDecoder(res.Body).Decode(&microData); err == nil && microData.Data.Screenshot.URL != "" {
				resURL := microData.Data.Screenshot.URL
				rw := microData.Data.Screenshot.Width
				rh := microData.Data.Screenshot.Height
				if rw == 0 {
					rw = w
				}
				if rh == 0 {
					rh = h
				}

				var rawBytes []byte
				if needRaw {
					if imgRes, err := client.Get(resURL); err == nil && imgRes.StatusCode == http.StatusOK {
						defer imgRes.Body.Close()
						rawBytes, _ = io.ReadAll(imgRes.Body)
					}
				}

				return &ProviderResult{
					ImageURL: resURL,
					Width:    rw,
					Height:   rh,
					FileSize: microData.Data.Screenshot.SizePretty,
					RawBytes: rawBytes,
				}, nil
			}
		}
	}

	thumWidth := 1920
	thumHeight := 1080
	if isMobile {
		thumWidth = 414
		thumHeight = 896
	}
	thumFull := ""
	if fullPage {
		thumFull = "fullpage/"
		thumHeight = 2400
	}
	thumURL := fmt.Sprintf("https://image.thum.io/get/width/%d/%s%s", thumWidth, thumFull, targetURL)

	reqThum, err := http.NewRequestWithContext(ctx, http.MethodGet, thumURL, nil)
	if err == nil {
		resThum, err := client.Do(reqThum)
		if err == nil && resThum.StatusCode == http.StatusOK {
			defer resThum.Body.Close()
			var rawBytes []byte
			if needRaw {
				rawBytes, _ = io.ReadAll(resThum.Body)
			}
			return &ProviderResult{
				ImageURL: thumURL,
				Width:    thumWidth,
				Height:   thumHeight,
				FileSize: "~150 KB",
				RawBytes: rawBytes,
			}, nil
		}
	}

	sshotRes := "1920x1080"
	ssWidth := 1920
	ssHeight := 1080
	if isMobile {
		sshotRes = "414x896"
		ssWidth = 414
		ssHeight = 896
	}
	sshotURL := fmt.Sprintf("https://mini.s-shot.ru/%s/PNG/%d/Z100/?%s", sshotRes, ssWidth, url.QueryEscape(targetURL))

	var rawBytes []byte
	if needRaw {
		if resSShot, err := client.Get(sshotURL); err == nil && resSShot.StatusCode == http.StatusOK {
			defer resSShot.Body.Close()
			rawBytes, _ = io.ReadAll(resSShot.Body)
		}
	}

	return &ProviderResult{
		ImageURL: sshotURL,
		Width:    ssWidth,
		Height:   ssHeight,
		FileSize: "~120 KB",
		RawBytes: rawBytes,
	}, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodPost && r.Method != http.MethodHead {
		respondJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Status: false,
			Error:  "Metode HTTP tidak didukung. Gunakan GET atau POST.",
		})
		return
	}

	targetURL := ""
	device := "desktop"
	fullPage := true
	raw := false

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		q := r.URL.Query()
		targetURL = strings.TrimSpace(q.Get("url"))
		if d := q.Get("device"); d != "" {
			device = d
		}
		if fp := q.Get("fullPage"); fp != "" {
			fullPage = fp != "false" && fp != "0"
		}
		raw = q.Get("raw") == "true"
	} else if r.Method == http.MethodPost {
		var body struct {
			URL      string `json:"url"`
			Device   string `json:"device"`
			FullPage any    `json:"fullPage"`
			Raw      any    `json:"raw"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			targetURL = strings.TrimSpace(body.URL)
			if body.Device != "" {
				device = body.Device
			}
			if body.FullPage != nil {
				s := fmt.Sprintf("%v", body.FullPage)
				fullPage = s != "false" && s != "0"
			}
			if bRaw, ok := body.Raw.(bool); ok {
				raw = bRaw
			}
		}
	}

	if targetURL == "" {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Status: false,
			Error:  "Parameter 'url' wajib diisi.",
		})
		return
	}

	targetURL = normalizeTargetURL(targetURL)
	normalizedDevice := "desktop"
	if strings.ToLower(device) == "mobile" {
		normalizedDevice = "mobile"
	}

	captured, err := captureWebsite(r.Context(), targetURL, normalizedDevice, fullPage, raw)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Status: false,
			Error:  "Gagal mengambil screenshot: " + err.Error(),
		})
		return
	}

	if raw {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		if len(captured.RawBytes) > 0 {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(captured.RawBytes)
			return
		}
		http.Redirect(w, r, captured.ImageURL, http.StatusFound)
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:3300"
	}

	selfImageURL := fmt.Sprintf("%s://%s/api/capture?url=%s&device=%s&fullPage=%v&raw=true",
		scheme, host, url.QueryEscape(targetURL), normalizedDevice, fullPage)

	respondJSON(w, http.StatusOK, APIResponse{
		Status: true,
		Code:   http.StatusOK,
		Result: &CaptureResult{
			TargetURL: targetURL,
			Device:    normalizedDevice,
			FullPage:  fullPage,
			Dimensions: map[string]int{
				"width":  captured.Width,
				"height": captured.Height,
			},
			FileSize: captured.FileSize,
			ImageURL: selfImageURL,
		},
	})
}
