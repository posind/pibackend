package phone

import "regexp"

func MaskPhoneNumber(phone string) string {
	re := regexp.MustCompile(`(\d{6})(\d{3})(\d{3,})`)
	return re.ReplaceAllString(phone, `$1xxx$3`)
}
