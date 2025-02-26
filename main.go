package main

import (
	_ "embed"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/qtgolang/SunnyNet/SunnyNet"
	sunnyhttp "github.com/qtgolang/SunnyNet/src/http"
	"github.com/qtgolang/SunnyNet/src/public"

	"wx_channel/pkg/argv"
	"wx_channel/pkg/certificate"
	"wx_channel/pkg/proxy"
	"wx_channel/pkg/util"
)

//go:embed certs/SunnyRoot.cer
var cert_data []byte

//go:embed lib/FileSaver.min.js
var file_saver_js []byte

//go:embed lib/jszip.min.js
var zip_js []byte

//go:embed inject/main.js
var main_js []byte

var Sunny = SunnyNet.NewSunny()
var version = "250215"
var v = "?t=" + version
var port = 2023

// 打印帮助信息
func print_usage() {
	fmt.Printf("Usage: wx_video_download [OPTION...]\n")
	fmt.Printf("Download WeChat video.\n\n")
	fmt.Printf("      --help                 display this help and exit\n")
	fmt.Printf("  -v, --version              output version information and exit\n")
	fmt.Printf("  -p, --port                 set proxy server network port\n")
	fmt.Printf("  -d, --dev                  set proxy server network device\n")
	os.Exit(0)
}

func main() {
	os_env := runtime.GOOS
	args := argv.ArgsToMap(os.Args) // 分解参数列表为Map
	if _, ok := args["help"]; ok {
		print_usage()
	} // 存在help则输出帮助信息并退出主程序
	if v, ok := args["v"]; ok { // 存在v则输出版本信息并退出主程序
		fmt.Printf("v%s %.0s\n", version, v)
		os.Exit(0)
	}
	if v, ok := args["version"]; ok { // 存在version则输出版本信息并退出主程序
		fmt.Printf("v%s %.0s\n", version, v)
		os.Exit(0)
	}
	// 设置参数默认值
	args["dev"] = argv.ArgsValue(args, "", "d", "dev")
	args["port"] = argv.ArgsValue(args, "", "p", "port")
	iport, errstr := strconv.Atoi(args["port"])
	if errstr != nil {
		args["port"] = strconv.Itoa(port) // 用户自定义值解析失败则使用默认端口
	} else {
		port = iport
	}

	delete(args, "p") // 删除冗余的参数p
	delete(args, "d") // 删除冗余的参数d

	signalChan := make(chan os.Signal, 1)
	// Notify the signal channel on SIGINT (Ctrl+C) and SIGTERM
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalChan
		fmt.Printf("\n正在关闭服务...%v\n\n", sig)
		if os_env == "darwin" {
			proxy.DisableProxyInMacOS(proxy.ProxySettings{
				Device:   args["dev"],
				Hostname: "127.0.0.1",
				Port:     args["port"],
			})
		}
		os.Exit(0)
	}()
	fmt.Printf("\nv" + version)
	fmt.Printf("\n问题反馈 https://github.com/ltaoo/wx_channels_download/issues\n")
	existing, err1 := certificate.CheckCertificate("SunnyNet")
	if err1 != nil {
		fmt.Printf("\nERROR %v\n", err1.Error())
		fmt.Printf("按 Ctrl+C 退出...\n")
		select {}
	}
	if !existing {
		fmt.Printf("\n\n正在安装证书...\n")
		err := certificate.InstallCertificate(cert_data)
		time.Sleep(3 * time.Second)
		if err != nil {
			fmt.Printf("\nERROR %v\n", err.Error())
			fmt.Printf("按 Ctrl+C 退出...\n")
			select {}
		}
	}
	Sunny.SetPort(port)
	Sunny.SetGoCallback(HttpCallback, nil, nil, nil)
	err := Sunny.Start().Error
	if err != nil {
		fmt.Printf("\nERROR %v\n", err.Error())
		fmt.Printf("按 Ctrl+C 退出...\n")
		select {}
	}
	proxy_server := fmt.Sprintf("127.0.0.1:%v", port)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   proxy_server,
			}),
		},
	}
	_, err3 := client.Get("https://sunny.io/")
	if err3 == nil {
		if os_env == "windows" {
			ok := Sunny.OpenDrive(false)
			if !ok {
				fmt.Printf("\nERROR 启动进程代理失败\n")
				fmt.Printf("按 Ctrl+C 退出...\n")
				select {}
			}
			Sunny.ProcessAddName("WeChatAppEx.exe")
		}
		if os_env == "darwin" {
			err := proxy.EnableProxyInMacOS(proxy.ProxySettings{
				Device:   args["dev"],
				Hostname: "127.0.0.1",
				Port:     args["port"],
			})
			if err != nil {
				fmt.Printf("\nERROR 设置代理失败 %v\n", err.Error())
				fmt.Printf("按 Ctrl+C 退出...\n")
				select {}
			}
		}
		color.Green(fmt.Sprintf("\n\n服务已正确启动，请打开需要下载的视频号页面进行下载"))
	} else {
		fmt.Println(fmt.Sprintf("\n\n您还未安装证书，请在浏览器打开 http://%v 并根据说明安装证书\n在安装完成后重新启动此程序即可\n", proxy_server))
	}
	fmt.Println("\n\n服务正在运行，按 Ctrl+C 退出...")
	select {}
}

type ChannelProfile struct {
	Title string `json:"title"`
}
type FrontendTip struct {
	Msg string `json:"msg"`
}

// 定义用户信息结构体
type UserProfile struct {
	Username    string    `json:"username,omitempty"`
	Nickname    string    `json:"nickname,omitempty"`
	Description string    `json:"description,omitempty"`
	Avatar      string    `json:"avatar,omitempty"`
	ID          string    `json:"id,omitempty"`
	CreateTime  int64     `json:"createtime,omitempty"`
	Videos      []VideoInfo `json:"videos,omitempty"`
	Contact     interface{} `json:"contact,omitempty"`
	Followers   int64     `json:"followers,omitempty"`
	Following   int64     `json:"following,omitempty"`
	ExtraInfo   map[string]interface{} `json:"extra_info,omitempty"`
}

type VideoInfo struct {
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	CoverURL  string `json:"coverUrl,omitempty"`
	URL       string `json:"url,omitempty"`
	Key       string `json:"key,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Duration  int64  `json:"duration,omitempty"`
	CreateTime int64 `json:"createtime,omitempty"`
}

// 全局变量用于存储用户信息
var userProfiles = make(map[string]*UserProfile)

// 保存用户信息到文件
func saveUserProfile(profile *UserProfile) {
	if profile == nil || (profile.ID == "" && profile.Username == "") {
		return
	}
	
	// 创建profiles目录（如果不存在）
	if _, err := os.Stat("profiles"); os.IsNotExist(err) {
		os.Mkdir("profiles", 0755)
	}
	
	// 生成文件名
	fileName := profile.Username
	if fileName == "" {
		fileName = profile.ID
	}
	if fileName == "" {
		fileName = profile.Nickname
	}
	if fileName == "" {
		fileName = fmt.Sprintf("unknown_%d", time.Now().Unix())
	}
	
	// 替换文件名中的特殊字符
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, ":", "_")
	fileName = strings.ReplaceAll(fileName, "?", "_")
	fileName = strings.ReplaceAll(fileName, "&", "_")
	fileName = strings.ReplaceAll(fileName, "=", "_")
	
	// 添加json扩展名
	fileName += ".json"
	
	// 保存到文件
	filePath := filepath.Join("profiles", fileName)
	profileJSON, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		fmt.Printf("序列化用户信息失败: %v\n", err)
		return
	}
	
	err = os.WriteFile(filePath, profileJSON, 0644)
	if err != nil {
		fmt.Printf("保存用户信息文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("\n已保存用户信息: %s\n", filePath)
}

// 解析JSON响应体并尝试提取用户信息
func extractUserProfileFromJSON(urlStr string, jsonData []byte) {
	// 如果JSON数据太短，可能不包含有用信息
	if len(jsonData) < 10 {
		return
	}
	
	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	
	path := parsedURL.Path
	
	// 先将JSON解析为通用结构
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		// 尝试解析为数组
		var dataArray []interface{}
		if err := json.Unmarshal(jsonData, &dataArray); err != nil {
			return
		}
		
		// 处理数组类型的响应
		if len(dataArray) > 0 {
			// 可能是视频列表或用户列表
			for _, item := range dataArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// 递归处理每个项目
					itemJSON, _ := json.Marshal(itemMap)
					extractUserProfileFromJSON(urlStr, itemJSON)
				}
			}
		}
		return
	}
	
	// 尝试提取用户信息
	var userProfile *UserProfile
	
	// 获取URL中的username参数
	username := extractUsernameFromURL(urlStr)
	
	// 处理不同的API路径和响应类型
	if strings.Contains(path, "/finder/profile") || 
		strings.Contains(path, "/feeds") || 
		strings.Contains(path, "/api/user") || 
		strings.Contains(urlStr, "username=") {
		
		// 从响应中提取用户信息
		userProfile = extractProfileFromData(data, username)
		
		// 如果找到用户信息，保存它
		if userProfile != nil && (userProfile.Username != "" || userProfile.ID != "") {
			// 如果没有username但有ID，尝试查找
			if userProfile.Username == "" && userProfile.ID != "" {
				for k, v := range userProfiles {
					if v.ID == userProfile.ID {
						userProfile.Username = k
						break
					}
				}
			}
			
			// 确保用户名不为空
			if userProfile.Username == "" && username != "" {
				userProfile.Username = username
			}
			
			// 如果仍然没有用户名，使用ID作为用户名
			if userProfile.Username == "" && userProfile.ID != "" {
				userProfile.Username = userProfile.ID
			}
			
			// 如果找到了用户名或ID
			if userProfile.Username != "" || userProfile.ID != "" {
				// 用作映射键的标识符
				identifier := userProfile.Username
				if identifier == "" {
					identifier = userProfile.ID
				}
				
				// 检查是否已存在该用户
				existingProfile, exists := userProfiles[identifier]
				if exists {
					// 合并信息
					mergeProfiles(existingProfile, userProfile)
					saveUserProfile(existingProfile)
					fmt.Printf("\n更新用户信息: %s (%s)\n", existingProfile.Nickname, identifier)
				} else {
					userProfiles[identifier] = userProfile
					saveUserProfile(userProfile)
					fmt.Printf("\n成功提取用户信息: %s (%s)\n", userProfile.Nickname, identifier)
				}
			}
		}
	} else if strings.Contains(path, "/finder/feed") {
		// 尝试提取视频列表信息
		videos := extractVideosFromFeed(data)
		if len(videos) > 0 {
			fmt.Printf("\n提取到 %d 个视频信息\n", len(videos))
			
			// 将视频信息添加到对应的用户
			// 先尝试查找URL中提到的用户
			if username != "" && userProfiles[username] != nil {
				for _, video := range videos {
					addVideoToProfile(userProfiles[username], video)
				}
				saveUserProfile(userProfiles[username])
			} else {
				// 尝试根据视频信息找到对应的用户
				for _, video := range videos {
					for _, profile := range userProfiles {
						if profile.ID == video.ID {
							addVideoToProfile(profile, video)
							saveUserProfile(profile)
							break
						}
					}
				}
			}
		}
	} else {
		// 对于其他类型的响应，尝试通用提取
		// 如果响应中有用户相关的字段如nickname, username, avatar等，可能是用户信息
		if hasUserFields(data) {
			userProfile = extractProfileFromData(data, username)
			if userProfile != nil && (userProfile.Username != "" || userProfile.ID != "") {
				// 与上面的保存逻辑相同
				identifier := userProfile.Username
				if identifier == "" {
					identifier = userProfile.ID
				}
				
				existingProfile, exists := userProfiles[identifier]
				if exists {
					mergeProfiles(existingProfile, userProfile)
					saveUserProfile(existingProfile)
				} else {
					userProfiles[identifier] = userProfile
					saveUserProfile(userProfile)
				}
			}
		}
	}
}

// 检查数据是否包含用户相关字段
func hasUserFields(data map[string]interface{}) bool {
	// 检查顶层字段
	userFields := []string{"nickname", "username", "avatar", "user_id", "user_name", "profile"}
	for _, field := range userFields {
		if _, ok := data[field]; ok {
			return true
		}
	}
	
	// 检查data字段
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		for _, field := range userFields {
			if _, ok := dataField[field]; ok {
				return true
			}
		}
		
		// 检查user字段
		if _, ok := dataField["user"].(map[string]interface{}); ok {
			return true
		}
		
		// 检查object字段
		if object, ok := dataField["object"].(map[string]interface{}); ok {
			if _, ok := object["nickname"]; ok {
				return true
			}
		}
	}
	
	return false
}

// 从不同数据结构中提取用户信息
func extractProfileFromData(data map[string]interface{}, username string) *UserProfile {
	profile := &UserProfile{
		Username: username,
		ExtraInfo: make(map[string]interface{}),
	}
	
	// 打印关键位置的数据结构，以便调试
	bytes, _ := json.MarshalIndent(data, "", "  ")
	if len(bytes) < 1000 {
		fmt.Printf("\n尝试提取用户信息，数据结构: %s\n", string(bytes))
	} else {
		fmt.Printf("\n尝试提取用户信息，数据结构较大: %d 字节\n", len(bytes))
	}
	
	// 尝试从data字段获取信息
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		// 从object字段获取信息
		if object, ok := dataField["object"].(map[string]interface{}); ok {
			if nickname, ok := object["nickname"].(string); ok {
				profile.Nickname = nickname
				fmt.Printf("从object字段提取到昵称: %s\n", nickname)
			}
			if id, ok := object["id"].(string); ok {
				profile.ID = id
				fmt.Printf("从object字段提取到ID: %s\n", id)
			}
			if createtime, ok := object["createtime"].(float64); ok {
				profile.CreateTime = int64(createtime)
			}
			if contact, ok := object["contact"]; ok {
				profile.Contact = contact
			}
			if desc, ok := object["objectDesc"].(map[string]interface{}); ok {
				if description, ok := desc["description"].(string); ok {
					profile.Description = description
					fmt.Printf("从object字段提取到描述: %s\n", description)
				}
			}
		}
		
		// 从user字段获取信息
		if userData, ok := dataField["user"].(map[string]interface{}); ok {
			if nickname, ok := userData["nickname"].(string); ok && profile.Nickname == "" {
				profile.Nickname = nickname
				fmt.Printf("从user字段提取到昵称: %s\n", nickname)
			}
			if id, ok := userData["id"].(string); ok && profile.ID == "" {
				profile.ID = id
				fmt.Printf("从user字段提取到ID: %s\n", id)
			}
			if avatar, ok := userData["avatar_url"].(string); ok && profile.Avatar == "" {
				profile.Avatar = avatar
			}
		}
		
		// 从author字段获取信息
		if author, ok := dataField["author"].(map[string]interface{}); ok {
			if avatar, ok := author["avatar_url"].(string); ok && profile.Avatar == "" {
				profile.Avatar = avatar
			}
			if nickname, ok := author["nickname"].(string); ok && profile.Nickname == "" {
				profile.Nickname = nickname
				fmt.Printf("从author字段提取到昵称: %s\n", nickname)
			}
		}
		
		// 从profile字段获取信息
		if profileField, ok := dataField["profile"].(map[string]interface{}); ok {
			if nickname, ok := profileField["nickname"].(string); ok && profile.Nickname == "" {
				profile.Nickname = nickname
				fmt.Printf("从profile字段提取到昵称: %s\n", nickname)
			}
			if avatar, ok := profileField["avatar"].(string); ok && profile.Avatar == "" {
				profile.Avatar = avatar
			}
			if desc, ok := profileField["desc"].(string); ok && profile.Description == "" {
				profile.Description = desc
			}
		}
		
		// 从统计信息中获取粉丝和关注数
		if statistics, ok := dataField["statistics"].(map[string]interface{}); ok {
			if followers, ok := statistics["follower_count"].(float64); ok {
				profile.Followers = int64(followers)
			} else if followers, ok := statistics["followers"].(float64); ok {
				profile.Followers = int64(followers)
			}
			
			if following, ok := statistics["following_count"].(float64); ok {
				profile.Following = int64(following)
			} else if following, ok := statistics["following"].(float64); ok {
				profile.Following = int64(following)
			}
		}
	}
	
	// 直接从顶层获取信息（用于某些API响应）
	if nickname, ok := data["nickname"].(string); ok && profile.Nickname == "" {
		profile.Nickname = nickname
		fmt.Printf("从顶层字段提取到昵称: %s\n", nickname)
	}
	if id, ok := data["id"].(string); ok && profile.ID == "" {
		profile.ID = id
		fmt.Printf("从顶层字段提取到ID: %s\n", id)
	}
	if username, ok := data["username"].(string); ok && profile.Username == "" {
		profile.Username = username
		fmt.Printf("从顶层字段提取到用户名: %s\n", username)
	}
	if avatar, ok := data["avatar"].(string); ok && profile.Avatar == "" {
		profile.Avatar = avatar
	}
	if createtime, ok := data["createtime"].(float64); ok && profile.CreateTime == 0 {
		profile.CreateTime = int64(createtime)
	}
	
	// 将未识别的数据存储到额外信息中
	for key, value := range data {
		if key != "data" && key != "code" && key != "msg" && key != "status" {
			profile.ExtraInfo[key] = value
		}
	}
	
	// 如果没有足够的信息，认为未提取成功
	if profile.Nickname == "" && profile.ID == "" {
		return nil
	}
	
	// 打印提取结果
	fmt.Printf("成功提取用户信息: 昵称=%s, ID=%s, 用户名=%s\n", 
		profile.Nickname, profile.ID, profile.Username)
	
	return profile
}

// 从feed响应中提取视频信息
func extractVideosFromFeed(data map[string]interface{}) []VideoInfo {
	var videos []VideoInfo
	
	// 尝试处理常见的数据结构
	if data, ok := data["data"].(map[string]interface{}); ok {
		if items, ok := data["items"].([]interface{}); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if object, ok := itemMap["object"].(map[string]interface{}); ok {
						video := VideoInfo{}
						
						if id, ok := object["id"].(string); ok {
							video.ID = id
						}
						
						if createtime, ok := object["createtime"].(float64); ok {
							video.CreateTime = int64(createtime)
						}
						
						if objectDesc, ok := object["objectDesc"].(map[string]interface{}); ok {
							if description, ok := objectDesc["description"].(string); ok {
								video.Title = description
							}
							
							if media, ok := objectDesc["media"].([]interface{}); ok && len(media) > 0 {
								if mediaItem, ok := media[0].(map[string]interface{}); ok {
									if coverUrl, ok := mediaItem["coverUrl"].(string); ok {
										video.CoverURL = coverUrl
									}
									
									if url, ok := mediaItem["url"].(string); ok {
										video.URL = url
										
										if urlToken, ok := mediaItem["urlToken"].(string); ok {
											video.URL += urlToken
										}
									}
									
									if decodeKey, ok := mediaItem["decodeKey"].(string); ok {
										video.Key = decodeKey
									}
									
									if fileSize, ok := mediaItem["fileSize"].(float64); ok {
										video.Size = int64(fileSize)
									}
									
									if spec, ok := mediaItem["spec"].([]interface{}); ok && len(spec) > 0 {
										if specItem, ok := spec[0].(map[string]interface{}); ok {
											if durationMs, ok := specItem["durationMs"].(float64); ok {
												video.Duration = int64(durationMs)
											}
										}
									}
								}
							}
						}
						
						if video.ID != "" {
							videos = append(videos, video)
						}
					}
				}
			}
		}
	}
	
	return videos
}

// 将视频信息添加到用户个人资料
func addVideoToProfile(profile *UserProfile, video VideoInfo) {
	// 检查视频是否已存在
	for i, existingVideo := range profile.Videos {
		if existingVideo.ID == video.ID {
			// 更新已存在的视频信息
			profile.Videos[i] = mergeVideoInfo(existingVideo, video)
			return
		}
	}
	
	// 添加新视频
	profile.Videos = append(profile.Videos, video)
}

// 合并两个用户个人资料
func mergeProfiles(dst, src *UserProfile) {
	if dst.Nickname == "" && src.Nickname != "" {
		dst.Nickname = src.Nickname
	}
	
	if dst.Description == "" && src.Description != "" {
		dst.Description = src.Description
	}
	
	if dst.Avatar == "" && src.Avatar != "" {
		dst.Avatar = src.Avatar
	}
	
	if dst.ID == "" && src.ID != "" {
		dst.ID = src.ID
	}
	
	if dst.CreateTime == 0 && src.CreateTime != 0 {
		dst.CreateTime = src.CreateTime
	}
	
	if dst.Contact == nil && src.Contact != nil {
		dst.Contact = src.Contact
	}
	
	if dst.Followers == 0 && src.Followers != 0 {
		dst.Followers = src.Followers
	}
	
	if dst.Following == 0 && src.Following != 0 {
		dst.Following = src.Following
	}
	
	// 合并视频信息
	for _, srcVideo := range src.Videos {
		addVideoToProfile(dst, srcVideo)
	}
	
	// 合并额外信息
	for k, v := range src.ExtraInfo {
		if _, exists := dst.ExtraInfo[k]; !exists {
			dst.ExtraInfo[k] = v
		}
	}
}

// 合并视频信息
func mergeVideoInfo(dst, src VideoInfo) VideoInfo {
	result := dst
	
	if result.Title == "" && src.Title != "" {
		result.Title = src.Title
	}
	
	if result.CoverURL == "" && src.CoverURL != "" {
		result.CoverURL = src.CoverURL
	}
	
	if result.URL == "" && src.URL != "" {
		result.URL = src.URL
	}
	
	if result.Key == "" && src.Key != "" {
		result.Key = src.Key
	}
	
	if result.Size == 0 && src.Size != 0 {
		result.Size = src.Size
	}
	
	if result.Duration == 0 && src.Duration != 0 {
		result.Duration = src.Duration
	}
	
	if result.CreateTime == 0 && src.CreateTime != 0 {
		result.CreateTime = src.CreateTime
	}
	
	return result
}

// 从URL中提取username
func extractUsernameFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	
	// 从查询参数中获取username
	queryValues := parsedURL.Query()
	username := queryValues.Get("username")
	
	// 如果查询参数中没有username，尝试从路径中获取
	if username == "" {
		pathSegments := strings.Split(parsedURL.Path, "/")
		for i, segment := range pathSegments {
			if segment == "profile" && i < len(pathSegments)-1 {
				username = pathSegments[i+1]
				break
			}
		}
	}
	
	return username
}

// 构建获取用户资料的API URL
func buildProfileAPIURL(username string) string {
	return fmt.Sprintf("https://channels.weixin.qq.com/api/profile.getProfile?username=%s", username)
}

// 构建获取用户视频列表的API URL
func buildFeedAPIURL(username string) string {
	return fmt.Sprintf("https://channels.weixin.qq.com/api/feeds.getFeedsProfile?username=%s&query_request_id=%s", 
		username, randomString(16))
}

// 主动请求用户资料和视频列表
func fetchUserProfile(username string) {
	if username == "" {
		return
	}
	
	// 获取用户资料
	profileURL := buildProfileAPIURL(username)
	profileReq, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return
	}
	
	// 设置请求头
	profileReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	profileReq.Header.Set("Accept", "application/json, text/plain, */*")
	
	// 发送请求
	client := &http.Client{}
	profileResp, err := client.Do(profileReq)
	if err != nil {
		return
	}
	defer profileResp.Body.Close()
	
	// 读取响应内容
	profileData, err := io.ReadAll(profileResp.Body)
	if err != nil {
		return
	}
	
	// 处理用户资料数据
	var profileJSON map[string]interface{}
	if err := json.Unmarshal(profileData, &profileJSON); err == nil {
		extractUserProfileFromJSON("profile.getProfile", profileJSON)
	}
	
	// 获取用户视频列表
	feedURL := buildFeedAPIURL(username)
	feedReq, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		return
	}
	
	// 设置请求头
	feedReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	feedReq.Header.Set("Accept", "application/json, text/plain, */*")
	
	// 发送请求
	feedResp, err := client.Do(feedReq)
	if err != nil {
		return
	}
	defer feedResp.Body.Close()
	
	// 读取响应内容
	feedData, err := io.ReadAll(feedResp.Body)
	if err != nil {
		return
	}
	
	// 处理用户视频列表数据
	var feedJSON map[string]interface{}
	if err := json.Unmarshal(feedData, &feedJSON); err == nil {
		extractUserProfileFromJSON("feeds.getFeedsProfile", feedJSON)
	}
}

// HttpCallback 处理HTTP请求和响应
func HttpCallback(sessID uint, isRequest bool, requestID uint, req *http.Request, resp *http.Response, data []byte) {
	// 使用URL()方法获取URL信息
	urlStr := req.URL.String()
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("解析URL错误:", err)
		return
	}
	host := parsedURL.Hostname()
	path := parsedURL.Path
	
	// 只拦截 channels.weixin.qq.com 的请求
	isTargetHost := host == "channels.weixin.qq.com"
	
	// 从URL中提取username，如果存在
	username := ""
	if isTargetHost {
		username = extractUsernameFromURL(urlStr)
		if username != "" {
			// 检查是否已经处理过该用户
			if _, exists := userProfiles[username]; !exists {
				// 创建新的用户配置文件
				userProfiles[username] = &UserProfile{
					Username: username,
					ExtraInfo: make(map[string]interface{}),
				}
				
				fmt.Printf("\n发现新用户: %s\n", username)
				
				// 异步获取用户资料，避免阻塞主线程
				go fetchUserProfile(username)
			}
		}
	}
	
	// 打印详细的请求信息
	if isRequest {
		if isTargetHost {
			fmt.Printf("\n===================== 请求 =====================\n")
			fmt.Printf("URL: %s\n", urlStr)
			
			// 打印请求头
			reqHeader := req.Header
			fmt.Println("请求头:")
			for k, v := range reqHeader {
				if len(v) > 0 {
					fmt.Printf("  %s: %s\n", k, v[0])
				}
			}
			
			// 打印请求体（如果存在）
			reqBody := req.Body
			if reqBody != nil {
				// 尝试格式化JSON
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, reqBody, "", "  "); err == nil {
					fmt.Printf("请求体(JSON):\n%s\n", prettyJSON.String())
				} else {
					// 如果不是JSON或解析失败，直接输出原始内容
					fmt.Printf("请求体:\n%s\n", string(reqBody))
				}
			}
			fmt.Printf("==============================================\n")
		}
		
		req.Header.Del("Accept-Encoding")
		if util.Includes(path, "jszip") {
			headers := sunnyhttp.Header{}
			headers.Set("Content-Type", "application/javascript")
			headers.Set("__debug", "local_file")
			Conn.StopRequest(200, zip_js, headers)
			return
		}
		if util.Includes(path, "FileSaver.min") {
			headers := sunnyhttp.Header{}
			headers.Set("Content-Type", "application/javascript")
			headers.Set("__debug", "local_file")
			Conn.StopRequest(200, file_saver_js, headers)
			return
		}
		if path == "/__wx_channels_api/profile" {
			var data ChannelProfile
			body := req.Body
			err := json.Unmarshal(body, &data)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Printf("\n打开了视频\n%s\n", data.Title)
			headers := sunnyhttp.Header{}
			headers.Set("Content-Type", "application/json")
			headers.Set("__debug", "fake_resp")
			Conn.StopRequest(200, []byte("{}"), headers)
			return
		}
		if path == "/__wx_channels_api/tip" {
			var data FrontendTip
			body := req.Body
			err := json.Unmarshal(body, &data)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Printf("[FRONTEND]%s\n", data.Msg)
			headers := sunnyhttp.Header{}
			headers.Set("Content-Type", "application/json")
			headers.Set("__debug", "fake_resp")
			Conn.StopRequest(200, []byte("{}"), headers)
			return
		}
	}
	if resp != nil {
		content_type := strings.ToLower(resp.Header.Get("content-type"))
		Body := resp.Body
		
		// 只拦截 channels.weixin.qq.com 的响应
		if isTargetHost {
			// 尝试从JSON响应中提取用户信息
			if strings.Contains(content_type, "application/json") && Body != nil && len(Body) > 0 {
				extractUserProfileFromJSON(urlStr, Body)
			}
			
			// 只打印API请求的响应，减少日志量
			if strings.Contains(path, "/api/") || 
				strings.Contains(path, "/finder/") || 
				strings.Contains(urlStr, "username=") {
				// 打印详细的响应信息
				fmt.Printf("\n===================== 响应 =====================\n")
				fmt.Printf("URL: %s\n", urlStr)
				
				// 打印响应状态码
				statusCode := resp.StatusCode
				fmt.Printf("状态码: %d\n", statusCode)
				
				// 不打印所有响应头，只打印内容类型和长度
				fmt.Printf("Content-Type: %s\n", content_type)
				if length := resp.Header.Get("content-length"); length != "" {
					fmt.Printf("Content-Length: %s\n", length)
				}
				
				// 只打印JSON格式的响应体
				if strings.Contains(content_type, "application/json") && Body != nil && len(Body) > 0 {
					// 尝试格式化JSON
					var prettyJSON bytes.Buffer
					if err := json.Indent(&prettyJSON, Body, "", "  "); err == nil {
						// 由于JSON可能很大，限制打印长度
						jsonStr := prettyJSON.String()
						if len(jsonStr) > 1000 {
							fmt.Printf("响应体(JSON, 已截断):\n%s...\n", jsonStr[:1000])
						} else {
							fmt.Printf("响应体(JSON):\n%s\n", jsonStr)
						}
					}
				} else if strings.Contains(content_type, "text/") {
					fmt.Printf("响应体: [文本内容] 长度: %d 字节\n", len(Body))
				} else {
					fmt.Printf("响应体: [二进制数据] 长度: %d 字节\n", len(Body))
				}
				fmt.Printf("==============================================\n")
			}
		}
		
		if content_type == "text/html; charset=utf-8" {
			html := string(Body)
			
			// 保存HTML内容到文件 - 只保存关键页面的HTML
			if isTargetHost && (strings.Contains(path, "/profile") || 
				strings.Contains(path, "/feed") || 
				strings.Contains(path, "/home")) {
				saveHTMLToFile(urlStr, Body)
			}
			
			script_reg1 := regexp.MustCompile(`src="([^"]{1,})\.js"`)
			html = script_reg1.ReplaceAllString(html, `src="$1.js`+v+`"`)
			script_reg2 := regexp.MustCompile(`href="([^"]{1,})\.js"`)
			html = script_reg2.ReplaceAllString(html, `href="$1.js`+v+`"`)
			Conn.GetResponseHeader().Set("__debug", "append_script")
			script2 := ""
			if host == "channels.weixin.qq.com" && (path == "/web/pages/feed" || path == "/web/pages/home") {
				script := fmt.Sprintf(`<script>%s</script>`, main_js)
				html = strings.Replace(html, "<head>", "<head>\n"+script+script2, 1)
				fmt.Println("1. 视频详情页 html 注入 js 成功")
				Conn.SetResponseBody([]byte(html))
				return
			}
			Conn.SetResponseBody([]byte(html))
			return
		}
		if content_type == "application/javascript" {
			content := string(Body)
			
			// 保存JavaScript内容到文件 - 只保存解密相关脚本
			if isTargetHost {
				if util.Includes(path, "/t/wx_fed/finder/web/web-finder/res/js/index.publish") ||
					util.Includes(path, "/t/wx_fed/finder/web/web-finder/res/js/virtual_svg-icons-register") ||
					util.Includes(path, "wasm_video_decode.js") {
					saveJSToFile(urlStr, Body)
					
					// 在控制台输出额外信息，说明这是加密相关的脚本
					if util.Includes(path, "/t/wx_fed/finder/web/web-finder/res/js/index.publish") {
						fmt.Println("\n找到视频播放解密脚本 - index.publish.js 已保存")
					}
					if util.Includes(path, "/t/wx_fed/finder/web/web-finder/res/js/virtual_svg-icons-register") {
						fmt.Println("\n找到视频信息处理脚本 - virtual_svg-icons-register.js 已保存")
					}
					if util.Includes(path, "wasm_video_decode.js") {
						fmt.Println("\n找到WASM视频解码脚本 - wasm_video_decode.js 已保存")
					}
				}
			}
			
			dep_reg := regexp.MustCompile(`"js/([^"]{1,})\.js"`)
			from_reg := regexp.MustCompile(`from {0,1}"([^"]{1,})\.js"`)
			lazy_import_reg := regexp.MustCompile(`import\("([^"]{1,})\.js"\)`)
			import_reg := regexp.MustCompile(`import {0,1}"([^"]{1,})\.js"`)
			content = from_reg.ReplaceAllString(content, `from"$1.js`+v+`"`)
			content = dep_reg.ReplaceAllString(content, `"js/$1.js`+v+`"`)
			content = lazy_import_reg.ReplaceAllString(content, `import("$1.js`+v+`")`)
			content = import_reg.ReplaceAllString(content, `import"$1.js`+v+`"`)
			Conn.GetResponseHeader().Set("__debug", "replace_script")

			if util.Includes(path, "/t/wx_fed/finder/web/web-finder/res/js/index.publish") {
				regexp1 := regexp.MustCompile(`this.sourceBuffer.appendBuffer\(h\),`)
				replaceStr1 := `(() => {
if (window.__wx_channels_store__) {
window.__wx_channels_store__.buffers.push(h);
}
})(),this.sourceBuffer.appendBuffer(h),`
				if regexp1.MatchString(content) {
					fmt.Println("2. 视频播放 js 修改成功")
				}
				content = regexp1.ReplaceAllString(content, replaceStr1)
				regexp2 := regexp.MustCompile(`if\(f.cmd===re.MAIN_THREAD_CMD.AUTO_CUT`)
				replaceStr2 := `if(f.cmd==="CUT"){
	if (window.__wx_channels_store__) {
	console.log("CUT", f, __wx_channels_store__.profile.key);
	window.__wx_channels_store__.keys[__wx_channels_store__.profile.key]=f.decryptor_array;
	}
}
if(f.cmd===re.MAIN_THREAD_CMD.AUTO_CUT`
				content = regexp2.ReplaceAllString(content, replaceStr2)
				Conn.SetResponseBody([]byte(content))
				return
			}
			if util.Includes(path, "/t/wx_fed/finder/web/web-finder/res/js/virtual_svg-icons-register") {
				regexp1 := regexp.MustCompile(`async finderGetCommentDetail\((\w+)\)\{return(.*?)\}async`)
				replaceStr1 := `async finderGetCommentDetail($1) {
					var feedResult = await$2;
					var data_object = feedResult.data.object;
					if (!data_object.objectDesc) {
						return feedResult;
					}
					var media = data_object.objectDesc.media[0];
					var profile = media.mediaType !== 4 ? {
						type: "picture",
						id: data_object.id,
						title: data_object.objectDesc.description,
						files: data_object.objectDesc.media,
						spec: [],
						contact: data_object.contact
					} : {
						type: "media",
						duration: media.spec[0].durationMs,
						spec: media.spec,
						title: data_object.objectDesc.description,
						coverUrl: media.coverUrl,
						url: media.url+media.urlToken,
						size: media.fileSize,
						key: media.decodeKey,
						id: data_object.id,
						nonce_id: data_object.objectNonceId,
						nickname: data_object.nickname,
						createtime: data_object.createtime,
						fileFormat: media.spec.map(o => o.fileFormat),
						contact: data_object.contact
					};
					fetch("/__wx_channels_api/profile", {
						method: "POST",
						headers: {
							"Content-Type": "application/json"
						},
						body: JSON.stringify(profile)
					});
					if (window.__wx_channels_store__) {
					__wx_channels_store__.profile = profile;
					window.__wx_channels_store__.profiles.push(profile);
					}
					return feedResult;
				}async`
				content = regexp1.ReplaceAllString(content, replaceStr1)
				Conn.SetResponseBody([]byte(content))
				return
			}
			Conn.SetResponseBody([]byte(content))
			return
		}
		Conn.SetResponseBody([]byte(Body))
		return
	}
}

// 辅助函数：取两个整数的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 保存HTML内容到文件
func saveHTMLToFile(urlStr string, content []byte) {
	// 创建html目录（如果不存在）
	if _, err := os.Stat("html"); os.IsNotExist(err) {
		os.Mkdir("html", 0755)
	}
	
	// 从URL生成文件名
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("解析URL错误:", err)
		return
	}
	
	// 生成文件名：将URL中的特殊字符替换为下划线
	fileName := strings.ReplaceAll(parsedURL.Path, "/", "_")
	if fileName == "" || fileName == "_" {
		fileName = "_index"
	}
	
	// 加上查询参数（如果存在）
	if parsedURL.RawQuery != "" {
		// 限制查询参数长度，避免文件名过长
		query := parsedURL.RawQuery
		if len(query) > 50 {
			query = query[:50]
		}
		fileName += "_" + strings.ReplaceAll(query, "&", "_")
	}
	
	// 添加时间戳，确保文件名唯一
	timestamp := time.Now().Format("20060102_150405")
	fileName += "_" + timestamp + ".html"
	
	// 保存文件
	filePath := filepath.Join("html", fileName)
	err = os.WriteFile(filePath, content, 0644)
	if err != nil {
		fmt.Printf("保存HTML文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("\n已保存HTML文件: %s\n", filePath)
}

// 保存JavaScript内容到文件
func saveJSToFile(urlStr string, content []byte) {
	// 创建js目录（如果不存在）
	if _, err := os.Stat("js"); os.IsNotExist(err) {
		os.Mkdir("js", 0755)
	}
	
	// 从URL生成文件名
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("解析URL错误:", err)
		return
	}
	
	// 生成文件名：将URL中的特殊字符替换为下划线
	fileName := strings.ReplaceAll(parsedURL.Path, "/", "_")
	if fileName == "" || fileName == "_" {
		fileName = "_index"
	}
	
	// 添加时间戳，确保文件名唯一
	timestamp := time.Now().Format("20060102_150405")
	fileName += "_" + timestamp + ".js"
	
	// 保存文件
	filePath := filepath.Join("js", fileName)
	err = os.WriteFile(filePath, content, 0644)
	if err != nil {
		fmt.Printf("保存JS文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("\n已保存JS文件: %s\n", filePath)
}

// 辅助函数：生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}