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
// 先写入 <filepath>.tmp，全部成功后原子替换，避免下载中断留下截断文件。
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

	tmpPath := filepath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}
	if err := os.Rename(tmpPath, filepath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
