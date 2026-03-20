package main

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type SentencesSimple struct {
	Id           *int    `json:"id"`
	Uuid         *string `json:"uuid"`
	Hitokoto     *string `json:"hitokoto"`
	SentenceType *string `json:"type"`
	From         *string `json:"from"`
	FromWho      *string `json:"from_who"`
	Creator      *string `json:"creator"`
	CreatorUid   *int    `json:"creator_uid"`
	Reviewer     *int    `json:"reviewer"`
	CommitFrom   *string `json:"commit_from"`
	CreatedAt    *string `json:"created_at"`
	Length       *int    `json:"length"`
}
type SentencesVersion struct {
	ProtocolVersion string                     `json:"protocol_version"`
	BundleVersion   string                     `json:"bundle_version"`
	UpDateAt        *string                    `json:"update_at"`
	Categories      SentencesVersionCategories `json:"categories"`
	Sentences       []SentencesCategories      `json:"sentences"`
}
type SentencesCategories struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	Desc     string  `json:"desc"`
	Key      string  `json:"key"`
	CreateAt *string `json:"create_at"`
	UpdateAt *string `json:"update_at"`
	Path     string  `json:"path"`
}
type SentencesVersionCategories struct {
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
}

func SentencesLoad(key string) ([]SentencesSimple, error) {
	var resp []SentencesSimple

	// 检查 sentences 目录是否存在
	if _, err := os.Stat("./sentences"); os.IsNotExist(err) {
		log.Println("无法读取句子包，正在尝试下载")
		baseUrl := "https://sentences-bundle.hitokoto.cn"
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		// 获取版本信息
		versionResp, err := client.Get(baseUrl + "/version.json")
		if err != nil {
			log.Println("无法下载句子包，请检查网络")
			return nil, err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatal("无法关闭响应体")
			}
		}(versionResp.Body)

		if versionResp.StatusCode != 200 {
			log.Println("无法下载句子包，服务器返回错误状态")
			return nil, err
		}

		versionContent, err := io.ReadAll(versionResp.Body)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		var sentencesVersion SentencesVersion
		err = json.Unmarshal(versionContent, &sentencesVersion)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		println(sentencesVersion.ProtocolVersion)

		// 创建 sentences 目录
		err = os.Mkdir("./sentences", 0755)
		if err != nil && !os.IsExist(err) {
			log.Fatal(err)
		}
		// 下载Version文件
		err = os.WriteFile("./sentences/version.json", versionContent, 0755)
		if err != nil {
			log.Fatal(err)
		}
		// 下载分类文件

		// 下载所有句子文件
		for _, path := range sentencesVersion.Sentences {
			println(baseUrl + "/" + path.Path)
			sentencesResp, err := client.Get(baseUrl + "/" + path.Path)
			if err != nil {
				log.Println("下载失败：" + path.Path)
				continue
			}

			if sentencesResp.StatusCode != 200 {
				log.Println("下载失败：" + path.Path)
				err := sentencesResp.Body.Close()
				if err != nil {
					return nil, err
				}
				continue
			}

			sentencesContent, err := io.ReadAll(sentencesResp.Body)
			if err != nil {
				err := sentencesResp.Body.Close()
				if err != nil {
					return nil, err
				}
				log.Fatal(err)
			}
			err = sentencesResp.Body.Close()
			if err != nil {
				return nil, err
			}

			err = os.WriteFile(filepath.Join("", path.Path), sentencesContent, 0755)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// 读取指定 key 的句子文件
	path := filepath.Join("./sentences", key+".json")
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("无法读取文件 %s: %v", path, err)
		return nil, err
	}

	err = json.Unmarshal(content, &resp)
	if err != nil {
		log.Printf("解析 JSON 失败：%v", err)
		return nil, err
	}

	return resp, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir("./wwwroot")).ServeHTTP(w, r)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PostFormValue("key")
	if key == "" {
		key = "a"
	}
	_, err := SentencesLoad(key)
	if err != nil {
		log.Fatal(err)
		return
	}
	var S SentencesSimple
	num := rand.Int()
	S.Id = &num
	print(num)
	payload, err := json.Marshal(S)
	if err != nil {
		log.Fatal(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(payload)
	if err != nil {
		log.Fatal(err)
		return
	}
	return
}

func main() {
	err := godotenv.Load(".env")
	log.Println("正在加载环境变量...")
	if err != nil {
		log.Fatalf("请在程序同级目录创建.env 文件：%s", err)
	}
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/v2", apiHandler)
	log.Println("正在启动服务...")
	info := os.Getenv("HOST") + ":" + os.Getenv("PORT")
	log.Println("服务器将在：" + info + "启动...")
	err = http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("Running on port " + os.Getenv("PORT") + "...")
}
