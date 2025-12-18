package networks

import (
	"fmt"
	"gocli/internal/utils"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func RunHTTPReqCommand(httpUrl, httpMethod, outputFile string, httpHeaders []string, params string) error {
	if !strings.HasPrefix(httpUrl, "http://") && !strings.HasPrefix(httpUrl, "https://") {
		httpUrl = "https://" + httpUrl
	}

	// 解析 URL
	baseURL, err := url.Parse(httpUrl)
	if err != nil {
		return fmt.Errorf("无效的 URL: %w", err)
	}

	var body io.Reader
	if httpMethod == "GET" {
		if params != "" {
			queryParams, err := utils.JSONToURLValues(params)
			if err != nil {
				return fmt.Errorf("解析参数失败: %w", err)
			}
			baseURL.RawQuery = queryParams.Encode()
		}
		body = nil
	} else {
		if params != "" {
			body = strings.NewReader(params)
		}
	}

	req, err := http.NewRequest(httpMethod, baseURL.String(), body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	for _, header := range httpHeaders {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if outputFile != "" {
		if err := saveToFile(outputFile, req, resp, &respBody); err != nil {
			return fmt.Errorf("保存响应失败: %w", err)
		}
	} else {
		if err := printToConsole(req, resp, &respBody); err != nil {
			return fmt.Errorf("打印响应失败: %w", err)
		}
	}
	return nil
}

func printToConsole(req *http.Request, resp *http.Response, body *[]byte) error {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("请求: %s %s\n", req.Method, req.URL.String())
	fmt.Printf("状态: %s\n", resp.Status)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("响应头:")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("响应体:")
	fmt.Println(string(*body))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

func saveToFile(filePath string, req *http.Request, resp *http.Response, body *[]byte) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	file, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开/创建文件失败: %w", err)
	}
	defer file.Close()

	if req.ProtoMajor == 2 && req.ProtoMinor == 0 {
		fmt.Fprintf(file, "HTTP/2.0 %s\n", resp.Status)
	} else if req.ProtoMajor == 1 && req.ProtoMinor == 1 {
		fmt.Fprintf(file, "HTTP/1.1 %s\n", resp.Status)
	} else if req.ProtoMajor == 1 && req.ProtoMinor == 0 {
		fmt.Fprintf(file, "HTTP/1.0 %s\n", resp.Status)
	}

	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(file, "%s: %s\n", key, value)
		}
	}
	fmt.Fprintln(file)

	_, err = file.Write(*body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✓ 响应已保存\n")
	fmt.Printf("  文件: %s\n", absPath)
	fmt.Printf("  状态: %s\n", resp.Status)
	fmt.Printf("  大小: %d 字节\n", len(*body))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}
