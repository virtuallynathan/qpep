package windivert

const (
	DIVERT_OK                  = 0
	DIVERT_ERROR_NOTINITILIZED = 1
	DIVERT_ERROR_ALREADY_INIT  = 2
	DIVERT_ERROR_FAILED        = 3
)

func InitializeWinDivertEngine(port, numThreads int) int {
	return DIVERT_OK
}

func CloseWinDivertEngine() int {
	return DIVERT_OK
}

func GetConnectionStateData(port int) (int, int, int, string, string) {
	return DIVERT_OK, -1, -1, "", ""
}

func EnableDiverterLogging(enable bool) {
	return
}
