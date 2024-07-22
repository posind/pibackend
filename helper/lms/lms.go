package lms

import (
	"github.com/gocroot/helper/atapi"
)

func GetTotalUser() (total int, err error) {

	url := "https://pamongdesa.id/webservice/user?page=1&perpage=1&search=&role%5B%5D=2&role%5B%5D=3&role%5B%5D=4&role%5B%5D=5&role%5B%5D=6&sub_position=&verification=&approval=&province=&regency=&district=&village=&start_date=&end_date=&statuslogin=%0A"
	_, res, err := atapi.GetWithBearer[Root]("", url)
	if err != nil {
		return
	}
	total = res.Data.Meta.Total
	return
}
