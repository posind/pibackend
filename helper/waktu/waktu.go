package waktu

import (
	"time"
)

func GetDateTimeJKTNow() (strdatetime string, err error) {
	// Tentukan lokasi untuk GMT+7
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return
	}
	// Dapatkan waktu sekarang dalam GMT+7
	now := time.Now().In(location)
	// Format waktu sesuai kebutuhan, contoh: "2006-01-02 15:04:05"
	strdatetime = now.Format("2006-01-02 15:04:05")
	return
}
