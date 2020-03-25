package monitor

import "os"

// 判断所给路径文件/文件夹是否存在
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDir(path string) (bool, error) {
	exist, err := Exists(path)
	if err != nil {
		return false, err
	}
	if !exist {
		if err :=os.MkdirAll(path, 0755); err!=nil{return false, err}
		os.Chmod(path, 0755)
	}

	return true, nil

}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}
