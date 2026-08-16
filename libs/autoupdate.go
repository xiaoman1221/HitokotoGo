package libs

import "log"

func AutoUpdate() bool {
	log.Println("开始自动更新句子流程")

	needUpdate, err := sentenceBundleNeedsUpdate()
	if err != nil {
		log.Printf("自动更新失败,无法检查远程版本: %v", err)
		return false
	}
	if !needUpdate {
		log.Println("句子包已是最新版本,跳过自动更新")
		return true
	}
	log.Println("检测到句子包有新版本,开始更新")

	if !CheckSentences() {
		log.Println("自动更新失败,将在下次任务周期中重试")
		return false
	}

	if err := ReloadSentences(); err != nil {
		log.Printf("自动更新后重新加载数据失败: %v", err)
		return false
	}

	localVersion, err := readLocalVersion()
	if err != nil {
		log.Printf("读取本地版本失败: %v", err)
		return false
	}
	log.Println("更新完成,当前版本: " + localVersion.BundleVersion)
	return true
}

func sentenceBundleNeedsUpdate() (bool, error) {
	localVersion, localErr := readLocalVersion()
	if localErr != nil {
		// 本地版本缺失或不可读，一律视为需要更新
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
