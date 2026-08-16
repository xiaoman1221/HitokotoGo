package libs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// httpClient 统一 HTTP 客户端，带超时，避免远程服务挂起时无限阻塞。
var httpClient = &http.Client{Timeout: 30 * time.Second}

// httpGet 读取远程文件内容（带状态码检查）。
func httpGet(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("关闭响应体失败: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// DownloadFile 下载远程文件到本地路径。
func DownloadFile(filepath, url string) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("关闭响应体失败: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Printf("关闭文件失败: %v", err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	return err
}
