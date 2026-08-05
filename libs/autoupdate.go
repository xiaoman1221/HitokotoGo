package libs

import (
	"HitokotoGo/entity"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func AutoUpdate() bool {
	log.Println("开始自动更新句子流程")

	needUpdate, err := sentenceBundleNeedsUpdate()
	if err != nil {
		log.Printf("自动更新失败，无法检查远程版本: %v", err)
		return false
	}
	if !needUpdate {
		log.Println("句子包已是最新版本，跳过自动更新")
		return true
	}
	log.Println("检测到句子包有新版本，开始更新")

	if !CheckSentences() {
		log.Printf("自动更新失败。将在下次任务周期中重试")
		return false
	}

	allSentences := LoadAllSentences("all")
	log.Printf("共加载 %d 条句子", len(allSentences))

	if refreshRedisAllData(allSentences) {
		localVersion, err := readLocalVersion()
		if err != nil {
			log.Printf("读取本地版本失败: %v", err)
			return false
		}
		log.Println("更新完成，已将所有REDIS数据刷写为最新版本，当前版本：" + localVersion.BundleVersion)
		return true
	}

	log.Println("无法连接到Redis，已刷新文件内存缓存")
	if err := restartCurrentProcess(); err != nil {
		log.Printf("自动重启失败: %v", err)
	}
	return true
}

func sentenceBundleNeedsUpdate() (bool, error) {
	localVersion, localErr := readLocalVersion()
	if localErr != nil {
		if os.IsNotExist(localErr) {
			return true, nil
		}
		return true, nil
	}

	remoteVersion, err := readRemoteVersion()
	if err != nil {
		return false, err
	}

	if localVersion.BundleVersion != remoteVersion.BundleVersion {
		return true, nil
	}
	if localVersion.UpdatedAt != remoteVersion.UpdatedAt {
		return true, nil
	}
	return false, nil
}

func readRemoteVersion() (*entity.V, error) {
	sentencesURL := os.Getenv("SENTENCES_URL")
	if sentencesURL == "" {
		return nil, fmt.Errorf("SENTENCES_URL 未配置")
	}

	resp, err := http.Get(sentencesURL + "/version.json")
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		if closeErr := body.Close(); closeErr != nil {
			log.Println("远程版本文件关闭失败")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("读取远程版本失败，状态码: %d", resp.StatusCode)
	}

	versionRemoteBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var remoteVersion entity.V
	if err := json.Unmarshal(versionRemoteBody, &remoteVersion); err != nil {
		return nil, err
	}

	return &remoteVersion, nil
}

func readLocalVersion() (*entity.V, error) {
	versionLocalBody, err := os.ReadFile("./data/version.json")
	if err != nil {
		return nil, err
	}
	var localVersion entity.V
	if err := json.Unmarshal(versionLocalBody, &localVersion); err != nil {
		return nil, err
	}
	return &localVersion, nil
}

func refreshRedisAllData(allSentences []entity.S) bool {
	if !InitRedis() {
		return false
	}
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		log.Printf("清空Redis数据失败: %v", err)
		return false
	}
	CacheAllSentences(allSentences)
	return true
}

func restartCurrentProcess() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(executable, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
