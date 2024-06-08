package normalize

import (
	"strconv"
	"strings"
)

// Fungsi untuk mengonversi angka menjadi representasi huruf
func NumberToAlphabet(num int) string {
	// Mengonversi angka menjadi string
	numStr := strconv.Itoa(num)

	// Array untuk menyimpan huruf-huruf hasil konversi
	var result []string

	// Iterasi melalui setiap karakter dalam string angka
	for _, char := range numStr {
		// Mengubah karakter digit menjadi huruf sesuai dengan urutan abjad
		letter := string('a' + (char - '0') - 1)
		result = append(result, letter)
	}

	// Gabungkan hasil ke dalam satu string
	return strings.Join(result, "")
}
