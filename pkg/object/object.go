package object

import (
	"WowjoyProject/WowjoyCallRadiant/global"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

//var token string

// 封装对象相关操作
type Object struct {
	Type            int
	Value           string
	Paths           []string
	AccessionNumber string
}

func NewObject(data global.RadiantData) *Object {
	// 获取文件查看影像路径
	paths := make([]string, 0)
	aceNum, paths := GetImagePath(data.ParamValue)
	return &Object{
		Type:            data.ParamType,
		Value:           data.ParamValue,
		Paths:           paths,
		AccessionNumber: aceNum,
	}
}

// 影像查看IMAGE_REVIEW
func (obj *Object) Image_Review() {
	global.Logger.Info("影像查看")
	switch global.ObjectSetting.CallRAType {
	case global.CallRadiantType_Share:
		SliceClear(&global.RadiantParam)
		global.RadiantParam = append(global.RadiantParam, obj.Paths...)
		CallRadiAnt(global.RadiantParam)
	case global.CallRadiantType_QR:
		SliceClear(&global.QRRadiantParam)
		global.QRRadiantParam = append(global.QRRadiantParam, obj.AccessionNumber)
		CallRadiAntQR(global.QRRadiantParam)
	}
}

// 追加影像查看APPEND_IMAGE_VIEW
func (obj *Object) Append_Image_View() {
	global.Logger.Info("追加影像查看")
	switch global.ObjectSetting.CallRAType {
	case global.CallRadiantType_Share:
		global.RadiantParam = append(global.RadiantParam, obj.Paths...)
		CallRadiAnt(global.RadiantParam)
	case global.CallRadiantType_QR:
		global.QRRadiantParam = append(global.QRRadiantParam, obj.AccessionNumber)
		CallRadiAntQR(global.QRRadiantParam)
	}
}

// 胶片直接打印
func (obj *Object) Film_Print() {
	global.Logger.Info("胶片直接打印")
	SliceClear(&global.RadiantParam)
	global.RadiantParam = append(global.RadiantParam, obj.Paths...)
	CallFilm(true, global.RadiantParam)
}

// 胶片追加打印
func (obj *Object) Append_Film_Print() {
	global.Logger.Info("胶片追加打印")
	SliceClear(&global.RadiantParam)
	global.RadiantParam = append(global.RadiantParam, obj.Paths...)
	CallFilm(false, global.RadiantParam)
}

// 发送数据给打印胶片程序
func CallFilm(flag bool, arg []string) {
	arg = RemoveDuplicate(arg)
	var tem_arg string
	tem_arg = global.ServerSetting.SendFilmCmd
	tem_arg += global.ServerSetting.SepCmd
	for _, k := range arg {
		tem_arg += k
		tem_arg += ","
	}
	tem_arg = tem_arg[:len(tem_arg)-1]

	address := "127.0.0.1" + ":" + global.ServerSetting.UDPFilmPort
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		global.Logger.Error(err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		global.Logger.Error("Dial.err", err)
		return
	}
	defer conn.Close()
	if flag {
		global.Logger.Info("发送清空打印胶片数据命令：", global.ServerSetting.CleanFilmData)
		conn.Write([]byte(global.ServerSetting.CleanFilmData))
	}
	global.Logger.Info("打印胶片发送命令：", string(tem_arg))
	conn.Write([]byte(tem_arg))
	// 发送显示窗口
	conn.Write([]byte("222$CMD_CON$1"))
}

// 解析UDP数据
func ParseUDPData(RecData string) {
	global.Logger.Info("开始解析UDP数据")
	if find := strings.Contains(RecData, global.ServerSetting.SepCon); !find {
		global.Logger.Error("UDP接收的数据不是有效数据，程序丢弃")
		return
	}
	indexSepCmd := strings.Index(RecData, global.ServerSetting.SepCmd)
	indexSepCon := strings.Index(RecData, global.ServerSetting.SepCon)

	objtype, _ := strconv.Atoi(RecData[indexSepCmd+len(global.ServerSetting.SepCmd) : indexSepCon])
	global.Logger.Debug("objtype: ", objtype)
	objvalue := RecData[indexSepCon+len(global.ServerSetting.SepCon):]
	global.Logger.Debug("objvalue: ", objvalue)

	// 判断是否开启下载模式
	switch global.ObjectSetting.CheckMode {
	case global.DOWN_MODE_ENABLE:
		global.Logger.Debug("开启下载模式: ", objvalue)
		// 调用服务端接口，下载数据
		CallDown(objvalue)
	default:
		global.Logger.Debug("未开启下载模式，直接打开影像...")
	}
	data := global.RadiantData{
		ParamType:  objtype,
		ParamValue: objvalue,
	}
	global.ObjectDataChan <- data
}

// 获取查看影像的路径
func GetImagePath(uid_enc string) (accessionNumber string, paths []string) {
	// 通过接口获取路径
	global.Logger.Debug("开始调用后台接口获取影像信息")
	url := global.ObjectSetting.IMAGE_URL
	url += uid_enc
	global.Logger.Debug("操作的URL: ", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		global.Logger.Error("调用获取影像信息接口失败：", err, uid_enc)
		return
	}
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		global.Logger.Error("Do Request got err: ", err)
		return
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		global.Logger.Error("Read resp.Body got err: ", err)
		return
	}
	//global.Logger.Debug("resp.Body: ", string(content))
	var result = make(map[string]interface{})
	err = json.Unmarshal(content, &result)
	if err != nil {
		global.Logger.Error("resp.Body: ", "错误")
		return
	}
	// 解析json
	if vCode, ok := result["code"]; ok {
		resultcode := vCode.(string)
		switch resultcode {
		case "0":
			global.Logger.Info("获取的接口正确，开始解析影像路径")
		default:
			global.Logger.Info("获取接口的数据错误", resultcode)
			return
		}
	}
	if vResult, ok := result["result"]; ok {
		if vResult != nil {
			resultMap := vResult.(map[string]interface{})
			if vaccessionNumber, ok := resultMap["accessionNumber"]; ok {
				accessionNumber = vaccessionNumber.(string)
			}
			if vSeriesList, ok := resultMap["seriesList"]; ok {
				if vSeriesList != nil {
					seriesList := vSeriesList.([]interface{})
					for _, seriesListItem := range seriesList {
						if seriesListItem != nil {
							seriesListMap := seriesListItem.(map[string]interface{})
							if vInstanceList, ok := seriesListMap["instanceList"]; ok {
								if vInstanceList != nil {
									instanceList := vInstanceList.([]interface{})
									for _, instanceListItem := range instanceList {
										if instanceListItem != nil {
											instanceListMap := instanceListItem.(map[string]interface{})
											var fileName, ip, sVirtualDir string
											if vfileName, ok := instanceListMap["fileName"]; ok {
												fileName = vfileName.(string)
												index := strings.LastIndex(fileName, "\\")
												fileName = fileName[:index]
											} else {
												continue
											}
											if vip, ok := instanceListMap["ip"]; ok {
												ip = vip.(string)
											} else {
												continue
											}
											if vsVirtualDir, ok := instanceListMap["sVirtualDir"]; ok {
												sVirtualDir = vsVirtualDir.(string)
											} else {
												continue
											}
											// 字符串拼接
											var buff bytes.Buffer
											buff.WriteString("\\\\")
											buff.WriteString(ip)
											buff.WriteString("\\")
											buff.WriteString(sVirtualDir)
											buff.WriteString("\\")
											buff.WriteString(fileName)
											paths = append(paths, buff.String())
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

// Slicce 清空操作
func SliceClear(s *[]string) {
	*s = (*s)[0:0]
}

// 调用Radiant客户端
func CallRadiAnt(arg []string) {
	exepath := global.ObjectSetting.RadiantPath
	arg = RemoveDuplicate(arg)
	tem_arg := make([]string, 0)
	tem_arg = append(tem_arg, "-cl")
	tem_arg = append(tem_arg, "-f")
	tem_arg = append(tem_arg, "-d")
	tem_arg = append(tem_arg, arg...)
	global.Logger.Info(exepath, " 参数是：", tem_arg)
	cmd := exec.Command(exepath, tem_arg...)
	if err := cmd.Run(); err != nil {
		global.Logger.Error("打开Radiant程序失败")
	}
}

// 调用Radiant客户端通过QR
func CallRadiAntQR(arg []string) {
	exepath := global.ObjectSetting.RadiantPath
	arg = RemoveDuplicate(arg)
	tem_arg := make([]string, 0)
	arglen := len(arg)
	tem_arg = append(tem_arg, "-cl")
	if arglen <= 1 {
		tem_arg = append(tem_arg, "-pstv")
		tem_arg = append(tem_arg, "00080050")
		tem_arg = append(tem_arg, arg...)
	} else {
		for _, value := range arg {
			tem_arg = append(tem_arg, "-pstv")
			tem_arg = append(tem_arg, "00080050")
			tem_arg = append(tem_arg, value)
		}
	}
	global.Logger.Info(exepath, " 参数是：", tem_arg)
	cmd := exec.Command(exepath, tem_arg...)
	if err := cmd.Run(); err != nil {
		global.Logger.Error("打开Radiant程序失败")
	}
}

// 去重操作
func RemoveDuplicate(arr []string) []string {
	resArr := make([]string, 0)
	tmpMap := make(map[string]interface{})
	for _, val := range arr {
		//判断主键为val的map是否存在
		if _, ok := tmpMap[val]; !ok {
			resArr = append(resArr, val)
			tmpMap[val] = nil
		}
	}
	return resArr
}

// 调用服务端下载接口
func CallDown(uid_enc string) {
	// 通过接口下载数据
	global.Logger.Debug("开始调用下载影像接口")
	url := global.ObjectSetting.Down_URL
	url += uid_enc
	global.Logger.Debug("操作的URL: ", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		global.Logger.Error("调用下载影像接口失败：", err, uid_enc)
		return
	}
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		global.Logger.Error("Do Request got err: ", err)
		return
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		global.Logger.Error("Read resp.Body got err: ", err)
		return
	}
	global.Logger.Debug("resp.Body: ", string(content))
}
