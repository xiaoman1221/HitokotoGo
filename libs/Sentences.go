package libs

import (
	"HitokotoGo/entity"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

func CheckSentences() bool {
	SentencesUrl := os.Getenv("SENTENCES_URL")
	log.Println("正在检查句子包")
	log.Println("正在检查data文件夹")
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		log.Println("data文件夹不存在,正在创建")
		err := os.Mkdir("data", 0755)
		if err != nil {
			log.Println("data文件夹创建失败,请检查权限")
			return false
		}
		log.Println("data文件夹创建成功")
	}
	log.Println("句子包检查更新中")
	// 检查文件是否存在
	if _, err := os.Stat("data/version.json"); os.IsNotExist(err) {
		log.Println("文件不存在")
		// 执行下载或其他操作
		err := DownloadFile("data/version.json", SentencesUrl+"/version.json")
		if err != nil {
			log.Println("文件下载失败,请检查权限")
			return false
		}
	} else if err != nil {
		log.Printf("检查文件失败: %v", err)
	} else {
		log.Println("文件存在")
	}
	versionLocal, err := os.ReadFile("data/version.json")
	if err != nil {
		log.Println("版本文件读取失败,请检查权限")
		return false
	}
	categoriesLocal, err := os.ReadFile("data/categories.json")
	if err != nil {
		log.Println("版本文件读取失败,请检查权限")
		return false
	}
	versionRemote, err := http.Get(SentencesUrl + "/version.json")
	if err != nil {
		log.Println("版本文件读取失败,请检查权限")
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("文件关闭失败")
		}
	}(versionRemote.Body)
	versionRemoteBody, VersionRBodyerr := io.ReadAll(versionRemote.Body)
	if VersionRBodyerr != nil {
		log.Println("版本文件解析失败,请检查权限")
	}

	var versionL entity.V
	var versionR entity.V
	var categories []entity.C
	var VersionRerr = json.Unmarshal(versionRemoteBody, &versionR)
	if VersionRerr != nil {
		log.Println("版本文件解析失败,请检查权限")
		return false
	}
	var versionLocalErr = json.Unmarshal(versionLocal, &versionL)
	if versionLocalErr != nil {
		log.Println("版本文件解析失败,请检查权限")
		return false
	}
	log.Println("版本文件读取成功")
	var categoriesLocalerr = json.Unmarshal(categoriesLocal, &categories)
	if categoriesLocalerr != nil {
		log.Println("分类文件解析失败,请检查权限")
		return false
	}
	log.Println("版本文件解析成功")
	log.Println("开始下载句子包")

	err = DownloadFile("./data/categories.json", SentencesUrl+versionL.Categories.Path)
	if err != nil {
		log.Println("句子分类下载失败,请检查权限")
		return false
	}
	if versionL.BundleVersion == versionR.BundleVersion || versionL.UpdatedAt == versionR.UpdatedAt {
		log.Println("句子包已是最新版本")
		return true
	}
	for _, version := range versionL.Sentences {
		log.Println("正在下载句子包: " + version.Path)
		if _, err := os.Stat("./data/sentences"); os.IsNotExist(err) {
			log.Println("sentences文件夹不存在,正在创建")
			err := os.Mkdir("./data/sentences", 0755)
			if err != nil {
				log.Println("sentences文件夹创建失败,请检查权限")
				return false
			}
			log.Println("sentences文件夹创建成功")
		}
		err := DownloadFile("./data/"+version.Path, SentencesUrl+version.Path)
		if err != nil {
			log.Println("句子包下载失败,请检查权限")
			return false
		}
	}
	return true
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
		return err
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
