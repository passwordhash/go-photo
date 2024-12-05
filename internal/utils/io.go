package utils

func IsAllowedExtension(extension string) bool {
	allowedExtension := map[string]struct{}{
		".jpg":  {},
		".jpeg": {},
		".png":  {},
	}

	if _, ok := allowedExtension[extension]; !ok {
		return false
	}

	return true
}
