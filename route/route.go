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
	//disable refresh token dulu
	//case method == "GET" && path == "/refresh/token":
	//	controller.GetNewToken(w, r)
	//akses data helpdesk layanan user
	case method == "GET" && path == "/data/user/helpdesk/all":
		controller.GetHelpdeskAll(w, r)
	case method == "GET" && path == "/data/user/helpdesk/masuk":
		controller.GetLatestHelpdeskMasuk(w, r)
	case method == "GET" && path == "/data/user/helpdesk/selesai":
		controller.GetLatestHelpdeskSelesai(w, r)

	//user data
	case method == "GET" && path == "/data/user":
		controller.GetDataUser(w, r)
	//mendapatkan data sent item
	case method == "GET" && at.URLParam(path, "/data/peserta/sent/:id"):
		controller.GetSentItem(w, r)
	//simpan feedback unsubs user
	case method == "POST" && path == "/data/peserta/unsubscribe": //resume atau risalah rapat dan feedback
		controller.PostUnsubscribe(w, r)
	//generate token linked device
	case method == "PUT" && path == "/data/user":
		controller.PutTokenDataUser(w, r)
	//mendapatkan data list nomor sender
	case method == "GET" && path == "/data/sender":
		controller.GetDataSenders(w, r)

	case method == "PUT" && path == "/data/user/task/doing":
		controller.PutTaskUser(w, r)
	case method == "GET" && path == "/data/user/task/done":
		controller.GetTaskDone(w, r)
	case method == "POST" && path == "/data/user/task/done":
		controller.PostTaskUser(w, r)
	case method == "GET" && path == "/data/pushrepo/kemarin":
		controller.GetYesterdayDistincWAGroup(w, r)
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
