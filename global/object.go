package global

var RadiantParam []string

const (
	IMAGE_REVIEW      int = 1 // 影像查看
	APPEND_IMAGE_VIEW int = 4 // 追加影像查看
)

const (
	DOWN_MODE_ENABLE   int = 1 //开启下载模式
	DOWN_MODE_UNENABLE int = 0 //关闭下载模式
)

var (
	ObjectDataChan chan RadiantData
)

type RadiantData struct {
	ParamType  int
	ParamValue string
}
