package utils

import "os"

func IsPhoto(extension string) bool {
	imageExt := map[string]struct{}{
		".jpg":  {},
		".jpeg": {},
		".png":  {},
		".svg":  {},
	}

	if _, ok := imageExt[extension]; !ok {
		return false
	}

	return true
}

func Exist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
