package application

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

// httpClient общий HTTP клиент для агентов application слоя
var httpClient = &http.Client{
	Timeout: 5 * time.Minute,
}

// httpPost выполняет POST запрос к OpenRouter API и возвращает тело + статус
func httpPost(ctx context.Context, url, apiKey string, body []byte) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://istok-agent-core.vercel.app")
	req.Header.Set("X-Title", "ИСТОК Агент")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}
