package route

import (
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/controller"
	"github.com/gocroot/helper/at"
)

func URL(w http.ResponseWriter, r *http.Request) {
	if config.SetAccessControlHeaders(w, r) {
		return // If it's a preflight request, return early.
	}
	config.SetEnv()

	var method, path string = r.Method, r.URL.Path
	switch {
	case method == "GET" && path == "/":
		controller.GetHome(w, r)
	//chat bot inbox
	case method == "POST" && at.URLParam(path, "/webhook/nomor/:nomorwa"):
		controller.PostInboxNomor(w, r)
	//masking list nmor official
	case method == "GET" && path == "/data/phone/all":
		controller.GetBotList(w, r)
	//akses data helpdesk layanan user
	case method == "GET" && path == "/data/user/helpdesk/all":
		controller.GetHelpdeskAll(w, r)
	case method == "GET" && path == "/data/user/helpdesk/masuk":
		controller.GetLatestHelpdeskMasuk(w, r)
	case method == "GET" && path == "/data/user/helpdesk/selesai":
		controller.GetLatestHelpdeskSelesai(w, r)
	//pamong desa data from api
	case method == "GET" && path == "/data/lms/user":
		controller.GetDataUserFromApi(w, r)
	//simpan testimoni dari pamong desa lms api
	case method == "POST" && path == "/data/lms/testi":
		controller.PostTestimoni(w, r)
		//get random 4 testi
	case method == "GET" && path == "/data/lms/random/testi":
		controller.GetRandomTesti4(w, r)
	//user data
	case method == "GET" && path == "/data/user":
		controller.GetDataUser(w, r)
	//mendapatkan data sent item
	case method == "GET" && at.URLParam(path, "/data/peserta/sent/:id"):
		controller.GetSentItem(w, r)
	//simpan feedback unsubs user
	case method == "POST" && path == "/data/peserta/unsubscribe":
		controller.PostUnsubscribe(w, r)
	//generate token linked device
	case method == "PUT" && path == "/data/user":
		controller.PutTokenDataUser(w, r)
	//Menambhahkan data nomor sender untuk broadcast
	case method == "PUT" && path == "/data/sender":
		controller.PutNomorBlast(w, r)
	//mendapatkan data list nomor sender untuk broadcast
	case method == "GET" && path == "/data/sender":
		controller.GetDataSenders(w, r)
	//mendapatkan data list nomor sender yang kena blokir dari broadcast
	case method == "GET" && path == "/data/blokir":
		controller.GetDataSendersTerblokir(w, r)
	//mendapatkan data rekap pengiriman wa blast
	case method == "GET" && path == "/data/rekap":
		controller.GetRekapBlast(w, r)
	//mendapatkan data faq
	case method == "GET" && path == "/data/faq":
		controller.GetFaq(w, r)
	case method == "POST" && path == "/data/faq":
		controller.PostFaq(w, r)
	case method == "PUT" && path == "/data/faq":
		controller.UpdateFaq(w, r)
	case method == "DELETE" && path == "/data/faq":
		controller.DeleteFaq(w, r)	
	
	//legacy
	case method == "PUT" && path == "/data/user/task/doing":
		controller.PutTaskUser(w, r)
	case method == "GET" && path == "/data/user/task/done":
		controller.GetTaskDone(w, r)
	case method == "POST" && path == "/data/user/task/done":
		controller.PostTaskUser(w, r)
	case method == "GET" && path == "/data/pushrepo/kemarin":
		controller.GetYesterdayDistincWAGroup(w, r)

	//helpdesk
	//mendapatkan data tiket
	case method == "GET" && at.URLParam(path, "/data/tiket/closed/:id"):
		controller.GetClosedTicket(w, r)
	//simpan feedback tiket user
	case method == "POST" && path == "/data/tiket/rate":
		controller.PostMasukanTiket(w, r)

	//disabel pendaftaran
	//case method == "POST" && path == "/data/user":
	//	controller.PostDataUser(w, r)
	//case method == "POST" && at.URLParam(path, "/data/user/wa/:nomorwa"):
	//	controller.PostDataUserFromWA(w, r)
	case method == "POST" && path == "/data/proyek":
		controller.PostDataProject(w, r)
	case method == "PUT" && path == "/data/proyek":
		controller.PutDataProject(w, r)
	case method == "DELETE" && path == "/data/proyek":
		controller.DeleteDataProject(w, r)
	case method == "GET" && path == "/data/proyek/anggota":
		controller.GetDataMemberProject(w, r)
	case method == "POST" && path == "/data/proyek/anggota":
		controller.PostDataMemberProject(w, r)
	case method == "POST" && path == "/approvebimbingan":
		controller.ApproveBimbinganbyPoin(w, r)
	case method == "DELETE" && path == "/data/proyek/anggota":
		controller.DeleteDataMemberProject(w, r)
	case method == "POST" && at.URLParam(path, "/webhook/github/:proyek"):
		controller.PostWebHookGithub(w, r)
	case method == "POST" && at.URLParam(path, "/webhook/gitlab/:proyek"):
		controller.PostWebHookGitlab(w, r)
	case method == "POST" && path == "/notif/ux/postlaporan":
		controller.PostLaporan(w, r)
	case method == "POST" && path == "/notif/ux/postfeedback":
		controller.PostFeedback(w, r)

	case method == "POST" && path == "/notif/ux/postmeeting":
		controller.PostMeeting(w, r)
	case method == "POST" && at.URLParam(path, "/notif/ux/postpresensi/:id"):
		controller.PostPresensi(w, r)
	case method == "POST" && at.URLParam(path, "/notif/ux/posttasklists/:id"):
		controller.PostTaskList(w, r)
	case method == "POST" && at.URLParam(path, "/webhook/nomor/:nomorwa"):
		controller.PostInboxNomor(w, r)
	// LMS
	case method == "GET" && path == "/lms/refresh/cookie":
		controller.RefreshLMSCookie(w, r)
	case method == "GET" && path == "/lms/count/user":
		controller.GetCountDocUser(w, r)
	// Google Auth
	case method == "POST" && path == "/auth/users":
		controller.Auth(w, r)
	case method == "POST" && path == "/auth/login":
		controller.GeneratePasswordHandler(w, r)
	case method == "POST" && path == "/auth/verify":
		controller.VerifyPasswordHandler(w, r)
	case method == "POST" && path == "/auth/resend":
		controller.ResendPasswordHandler(w, r)
	// Google Auth
	default:
		controller.NotFound(w, r)
	}
}
