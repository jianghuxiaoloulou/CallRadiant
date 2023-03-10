package setting

type ServerSettingS struct {
	RunMode       string
	UDPPort       string
	SepCmd        string
	SepCon        string
	UDPFilmPort   string
	CleanFilmData string
	SendFilmCmd   string
}

type GeneralSettingS struct {
	LogSavePath string
	LogFileName string
	LogFileExt  string
	LogMaxSize  int
	LogMaxAge   int
	MaxThreads  int
	MaxTasks    int
}

type ObjectSettingS struct {
	IMAGE_URL   string
	RadiantPath string
	CheckMode   int
	Down_URL    string
	CallRAType  int
}

func (s *Setting) ReadSection(k string, v interface{}) error {
	err := s.vp.UnmarshalKey(k, v)
	if err != nil {
		return err
	}
	return nil
}
