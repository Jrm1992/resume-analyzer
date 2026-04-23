//go:build integration

package integration

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jose/resume-analyzer/internal/config"
	apphttp "github.com/jose/resume-analyzer/internal/http"
	"github.com/jose/resume-analyzer/internal/jobs"
	"github.com/jose/resume-analyzer/internal/llm"
)

func TestEndToEnd_AgainstRealProvider(t *testing.T) {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		t.Skip("set LLM_API_KEY to run integration test")
	}
	model := os.Getenv("LLM_MODEL")
	if model == "" {
		t.Skip("set LLM_MODEL to run integration test")
	}
	base := os.Getenv("LLM_BASE_URL")
	if base == "" {
		base = "https://api.openai.com/v1"
	}

	cfg := &config.Config{
		MaxPDFBytes:   10 << 20,
		LLMTimeout:    120 * time.Second,
		Workers:       1,
		QueueCapacity: 4,
		JobTTL:        time.Hour,
		LLMBaseURL:    base,
		LLMAPIKey:     apiKey,
		LLMModel:      model,
		LLMMaxTokens:  4000,
	}
	tpl, err := apphttp.LoadTemplates()
	if err != nil {
		t.Fatal(err)
	}
	store := jobs.NewStore()
	queue := jobs.NewQueue(cfg.Workers, cfg.QueueCapacity)
	client := &llm.Client{
		BaseURL:   cfg.LLMBaseURL,
		APIKey:    cfg.LLMAPIKey,
		Model:     cfg.LLMModel,
		MaxTokens: cfg.LLMMaxTokens,
		Timeout:   cfg.LLMTimeout,
		HTTP:      &http.Client{Timeout: cfg.LLMTimeout + 5*time.Second},
	}
	srv := &apphttp.Server{Config: cfg, Templates: tpl, Store: store, Queue: queue, Analyzer: client}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	queue.Start(ctx, srv.JobHandler)

	resumePDF := samplePDF(t)
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("resume", "resume.pdf")
	_, _ = fw.Write(resumePDF)
	_ = mw.WriteField("jd", "We're hiring a senior Go engineer. Required: Go, PostgreSQL, Kubernetes.")
	mw.Close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/analyze", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	srv.Router().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("analyze status = %d: %s", rec.Code, rec.Body.String())
	}
	jobID := extractJobID(t, rec.Body.String())

	deadline := time.Now().Add(150 * time.Second)
	for time.Now().Before(deadline) {
		r2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/jobs/"+jobID, nil)
		srv.Router().ServeHTTP(r2, req2)
		if strings.Contains(r2.Body.String(), "Match score") {
			r3 := httptest.NewRecorder()
			req3 := httptest.NewRequest("GET", "/jobs/"+jobID+"/pdf", nil)
			srv.Router().ServeHTTP(r3, req3)
			if r3.Code != 200 {
				t.Fatalf("pdf status = %d", r3.Code)
			}
			data, _ := io.ReadAll(r3.Body)
			if !bytes.HasPrefix(data, []byte("%PDF-")) {
				t.Fatal("not a PDF")
			}
			return
		}
		if strings.Contains(r2.Body.String(), "Analysis failed") {
			t.Fatalf("job failed: %s", r2.Body.String())
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("timeout waiting for done")
}

func extractJobID(t *testing.T, body string) string {
	t.Helper()
	const marker = `hx-get="/jobs/`
	i := strings.Index(body, marker)
	rest := body[i+len(marker):]
	end := strings.Index(rest, `"`)
	return rest[:end]
}

func samplePDF(t *testing.T) []byte {
	t.Helper()
	content := "BT /F1 12 Tf 72 720 Td (Jane Doe - Senior Backend Engineer - 8 years Go PostgreSQL) Tj ET"
	stream := []byte(content)
	var pdfBuf bytes.Buffer
	pdfBuf.WriteString("%PDF-1.4\n")
	objs := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n",
		"4 0 obj\n<< /Length " + itoa3(len(stream)) + " >>\nstream\n" + string(stream) + "\nendstream\nendobj\n",
		"5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
	}
	offsets := []int{0}
	for _, o := range objs {
		offsets = append(offsets, pdfBuf.Len())
		pdfBuf.WriteString(o)
	}
	xref := pdfBuf.Len()
	pdfBuf.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for _, off := range offsets[1:] {
		s := itoa3(off)
		for len(s) < 10 {
			s = "0" + s
		}
		pdfBuf.WriteString(s + " 00000 n \n")
	}
	pdfBuf.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n" + itoa3(xref) + "\n%%EOF\n")
	return pdfBuf.Bytes()
}
func itoa3(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
