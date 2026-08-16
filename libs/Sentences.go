package libs

import (
	"HitokotoGo/entity"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CheckSentences() bool {
	const (
		dataDir      = "data"
		sentencesDir = "data/sentences"
		versionFile  = "data/version.json"
		dirPerm      = 0755
	)

	ensureDir := func(path string) error {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Printf("%s 不存在,正在创建", path)
			if err := os.MkdirAll(path, dirPerm); err != nil {
				return err
			}
			log.Printf("%s 创建成功", path)
			return nil
		} else if err != nil {
			return err
		}
		return nil
	}

	// cleanRelPath 去掉版本文件里 "./" 之类的前缀，统一拼接本地路径与远程 URL。
	cleanRelPath := func(p string) string {
		p = strings.TrimPrefix(p, "./")
		return strings.TrimPrefix(p, "/")
	}

	downloadAll := func(remoteVersion entity.V, sentencesURL string) bool {
		if err := ensureDir(dataDir); err != nil {
			log.Println("data文件夹创建失败,请检查权限")
			return false
		}
		if err := ensureDir(sentencesDir); err != nil {
			log.Println("sentences文件夹创建失败,请检查权限")
			return false
		}

		log.Println("开始下载分类文件")
		catsPath := cleanRelPath(remoteVersion.Categories.Path)
		if err := DownloadFile(filepath.Join(dataDir, catsPath), sentencesURL+"/"+catsPath); err != nil {
			log.Printf("句子分类下载失败: %v", err)
			return false
		}

		for _, sentence := range remoteVersion.Sentences {
			log.Println("正在下载句子包: " + sentence.Path)
			relPath := cleanRelPath(sentence.Path)
			if err := DownloadFile(filepath.Join(dataDir, relPath), sentencesURL+"/"+relPath); err != nil {
				log.Printf("句子包下载失败: %v", err)
				return false
			}
		}

		if err := DownloadFile(versionFile, sentencesURL+"/version.json"); err != nil {
			log.Printf("版本文件更新失败: %v", err)
			return false
		}

		return true
	}

	sentencesURL := os.Getenv("SENTENCES_URL")
	if sentencesURL == "" {
		log.Println("SENTENCES_URL 未配置")
		return false
	}

	log.Println("正在检查句子包")
	log.Println("正在检查data文件夹")

	remoteVersion, err := readRemoteVersion()
	if err != nil {
		log.Printf("远程版本文件读取失败: %v", err)
		return false
	}
	log.Println("远程版本文件读取成功")

	dataInfo, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		log.Println("data文件夹不存在,开始下载全部文件")
		return downloadAll(*remoteVersion, sentencesURL)
	}
	if err != nil {
		log.Printf("data文件夹检查失败: %v", err)
		return false
	}
	if !dataInfo.IsDir() {
		log.Println("data路径不是文件夹")
		return false
	}

	localVersion, err := readLocalVersion()
	if os.IsNotExist(err) {
		log.Println("version.json 不存在,开始下载全部文件")
		return downloadAll(*remoteVersion, sentencesURL)
	}
	if err != nil {
		log.Printf("版本文件读取失败: %v", err)
		return false
	}

	if localVersion.BundleVersion == remoteVersion.BundleVersion && localVersion.UpdatedAt == remoteVersion.UpdatedAt {
		log.Println("句子包已是最新版本")
		return true
	}

	log.Println("本地版本与远程版本不一致,开始覆盖下载全部文件")
	return downloadAll(*remoteVersion, sentencesURL)
}

// HasLocalData 判断本地是否已有可用的句子数据（用于远程检查失败时降级启动）。
func HasLocalData() bool {
	if _, err := os.Stat("data/version.json"); err != nil {
		return false
	}
	info, err := os.Stat("data/sentences")
	if err != nil || !info.IsDir() {
		return false
	}
	entries, err := os.ReadDir("data/sentences")
	if err != nil || len(entries) == 0 {
		return false
	}
	return true
}
