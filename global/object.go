package global

var RadiantParam []string
var QRRadiantParam []string

const (
	CallRadiantType_Share int = 1 // 共享方式访问文件
	CallRadiantType_QR    int = 2 // QR方式访问文件
)

const (
	IMAGE_REVIEW      int = 1  // 影像查看
	FILM_PRINT        int = 3  // 胶片直接打印
	APPEND_IMAGE_VIEW int = 4  // 追加影像查看
	APPEND_FILM_PRINT int = 18 // 胶片追加打印
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
