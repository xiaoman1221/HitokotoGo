package libs

import (
	"HitokotoGo/entity"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func CheckSentences() bool {
	const (
		dataDir                    = "data"
		sentencesDir               = "data/sentences"
		versionFile                = "data/version.json"
		categoriesFile             = "data/categories.json"
		dirPerm        os.FileMode = 0755
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

	readRemoteVersion := func(url string) ([]byte, error) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer func(body io.ReadCloser) {
			if closeErr := body.Close(); closeErr != nil {
				log.Println("文件关闭失败")
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		return io.ReadAll(resp.Body)
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
		if err := DownloadFile(categoriesFile, sentencesURL+remoteVersion.Categories.Path); err != nil {
			log.Println("句子分类下载失败,请检查权限")
			return false
		}

		for _, sentence := range remoteVersion.Sentences {
			log.Println("正在下载句子包: " + sentence.Path)
			if err := DownloadFile(dataDir+"/"+sentence.Path, sentencesURL+sentence.Path); err != nil {
				log.Println("句子包下载失败,请检查权限")
				return false
			}
		}

		if err := DownloadFile(versionFile, sentencesURL+"/version.json"); err != nil {
			log.Println("版本文件更新失败,请检查权限")
			return false
		}

		return true
	}

	sentencesURL := os.Getenv("SENTENCES_URL")
	log.Println("正在检查句子包")
	log.Println("正在检查data文件夹")

	versionRemoteBody, err := readRemoteVersion(sentencesURL + "/version.json")
	if err != nil {
		log.Println("远程版本文件读取失败,请检查权限")
		return false
	}

	var remoteVersion entity.V
	if err := json.Unmarshal(versionRemoteBody, &remoteVersion); err != nil {
		log.Println("远程版本文件解析失败,请检查权限")
		return false
	}

	log.Println("远程版本文件读取成功")
	log.Println("远程版本文件解析成功")

	dataInfo, err := os.Stat(dataDir)
	if os.IsNotExist(err) {
		log.Println("data文件夹不存在,开始下载全部文件")
		return downloadAll(remoteVersion, sentencesURL)
	}
	if err != nil {
		log.Printf("data文件夹检查失败: %v", err)
		return false
	}
	if !dataInfo.IsDir() {
		log.Println("data路径不是文件夹")
		return false
	}

	versionLocalBody, err := os.ReadFile(versionFile)
	if os.IsNotExist(err) {
		log.Println("version.json 不存在,开始下载全部文件")
		return downloadAll(remoteVersion, sentencesURL)
	}
	if err != nil {
		log.Println("版本文件读取失败,请检查权限")
		return false
	}

	var localVersion entity.V
	if err := json.Unmarshal(versionLocalBody, &localVersion); err != nil {
		log.Println("本地版本文件解析失败,开始重新下载全部文件")
		return downloadAll(remoteVersion, sentencesURL)
	}

	log.Println("本地版本文件读取成功")
	log.Println("本地版本文件解析成功")

	if localVersion.BundleVersion == remoteVersion.BundleVersion && localVersion.UpdatedAt == remoteVersion.UpdatedAt {
		log.Println("句子包已是最新版本")
		return true
	}

	log.Println("本地版本与远程版本不一致,开始覆盖下载全部文件")
	return downloadAll(remoteVersion, sentencesURL)
}
func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("文件关闭失败")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Println("文件关闭失败")
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}
