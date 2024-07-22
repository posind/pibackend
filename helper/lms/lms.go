package lms

import (
	"fmt"
	"io"
	"net/http"
)

func GetFirst() {

	url := "https://pamongdesa.id/webservice/user?page=1&perpage=1&search=&role%5B%5D=2&role%5B%5D=3&role%5B%5D=4&role%5B%5D=5&role%5B%5D=6&sub_position=&verification=&approval=&province=&regency=&district=&village=&start_date=&end_date=&statuslogin=%0A"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "••••••")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
