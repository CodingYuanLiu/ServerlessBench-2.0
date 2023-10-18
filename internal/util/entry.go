package util

func GetFuncInfoSlice(appInfo map[string]string) (ret []string) {
	var srcDir string
	var reqPath string
	var memSize string
	var retSlice []string

	for name, value := range appInfo {
		switch name {
		case "src":
			srcDir = value
		case "req":
			reqPath = value
		case "memory":
			memSize = value
		}
	}
	retSlice = append(retSlice, srcDir)
	retSlice = append(retSlice, reqPath)
	retSlice = append(retSlice, memSize)
	return retSlice
}
