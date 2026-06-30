package mcpstdio

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const contentLengthHeader = "content-length"

type Proxy struct {
	Endpoint  string
	Token     string
	HTTP      *http.Client
	UserAgent string
}

func NewProxy(instance, token string) *Proxy {
	return &Proxy{
		Endpoint:  strings.TrimRight(instance, "/") + "/mcp",
		Token:     token,
		HTTP:      &http.Client{Timeout: 60 * time.Second},
		UserAgent: "openpost-mcp/0.1.0",
	}
}

func (p *Proxy) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)
	for {
		frame, err := ReadFrame(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		resp, err := p.Forward(ctx, frame)
		if err != nil {
			resp = jsonRPCError(frame, err)
		}
		if len(resp) == 0 {
			continue
		}
		if err := WriteFrame(out, resp); err != nil {
			return err
		}
	}
}

func (p *Proxy) Forward(ctx context.Context, frame []byte) ([]byte, error) {
	if strings.TrimSpace(p.Endpoint) == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if _, err := url.ParseRequestURI(p.Endpoint); err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", p.UserAgent)
	if p.Token != "" {
		req.Header.Set("Authorization", "Bearer "+p.Token)
	}

	client := p.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return nil, fmt.Errorf("remote MCP returned HTTP %d: %s", resp.StatusCode, message)
	}
	return body, nil
}

func ReadFrame(r *bufio.Reader) ([]byte, error) {
	headers := map[string]string{}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF && strings.TrimSpace(line) == "" {
				return nil, io.EOF
			}
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid MCP header line %q", line)
		}
		headers[strings.ToLower(strings.TrimSpace(name))] = strings.TrimSpace(value)
	}

	rawLength := headers[contentLengthHeader]
	if rawLength == "" {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	length, err := strconv.Atoi(rawLength)
	if err != nil || length < 0 {
		return nil, fmt.Errorf("invalid Content-Length %q", rawLength)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

func WriteFrame(w io.Writer, body []byte) error {
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err := w.Write(body)
	return err
}

func jsonRPCError(request []byte, cause error) []byte {
	var in struct {
		ID any `json:"id,omitempty"`
	}
	_ = json.Unmarshal(request, &in)
	payload := struct {
		JSONRPC string `json:"jsonrpc"`
		ID      any    `json:"id,omitempty"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{
		JSONRPC: "2.0",
		ID:      in.ID,
	}
	payload.Error.Code = -32000
	payload.Error.Message = cause.Error()
	out, err := json.Marshal(payload)
	if err != nil {
		return []byte(`{"jsonrpc":"2.0","error":{"code":-32000,"message":"proxy error"}}`)
	}
	return out
}
