package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PostWebHookGithub(respw http.ResponseWriter, req *http.Request) {
	var resp model.Response
	prj, err := atdb.GetOneDoc[model.Project](config.Mongoconn, "project", primitive.M{"name": helper.GetParam(req)})
	if err != nil {
		resp.Info = "Tidak terdaftar"
		resp.Response = err.Error()
		helper.WriteJSON(respw, http.StatusUnavailableForLegalReasons, resp)
		return
	}
	hook, err := github.New(github.Options.Secret(prj.Secret))
	if err != nil {
		resp.Info = "Tidak berhak"
		resp.Response = err.Error()
		helper.WriteJSON(respw, http.StatusUnauthorized, resp)
		return
	}
	payload, err := hook.Parse(req, github.PushEvent, github.PingEvent)
	if err != nil {
		resp.Info = "Tidak ada Push"
		resp.Response = err.Error()
		helper.WriteJSON(respw, http.StatusBadRequest, resp)
		return
	}
	switch pyl := payload.(type) {
	case github.PingPayload:
		resp.Response = prj.Description
		resp.Info = prj.Name
		resp.Status = prj.Owner.Name
		helper.WriteJSON(respw, http.StatusOK, resp)
		return
	case github.PushPayload:
		var komsg, msg string
		for i, komit := range pyl.Commits {
			kommsg := strings.TrimSpace(komit.Message)
			appd := strconv.Itoa(i+1) + ". " + kommsg + "\n_" + komit.Author.Name + "_\n"
			dokcommit := model.PushReport{
				ProjectName: prj.Name,
				Project:     prj,
				Username:    komit.Author.Username,
				Email:       komit.Author.Email,
				Repo:        pyl.Repository.URL,
				Ref:         pyl.Ref,
				Message:     kommsg,
			}
			if (prj.Owner.Email == komit.Author.Email) || (prj.Owner.GithubUsername == komit.Author.Username) {
				dokcommit.User = prj.Owner
			} else {
				var member *model.Userdomyikado
				member, err := getMemberByAttributeInProject(prj, "githubusername", komit.Author.Username)
				if err != nil {
					member, err = getMemberByAttributeInProject(prj, "email", komit.Author.Email)
					if err != nil {
						resp.Location = komit.Author.Email + " | " + komit.Author.Username
						resp.Info = "Username dan Email di GitHub tidak terdaftar"
						resp.Response = err.Error()
						helper.WriteJSON(respw, http.StatusLocked, resp)
						return
					}
				}
				dokcommit.User = *member
			}
			_, err = atdb.InsertOneDoc(config.Mongoconn, "pushrepo", dokcommit)
			if err != nil {
				resp.Info = "Tidak masuk ke database"
				resp.Response = err.Error()
				helper.WriteJSON(respw, http.StatusExpectationFailed, resp)
				return
			}
			komsg += appd
		}
		msg = pyl.Pusher.Name + "\n" + pyl.Sender.Login + "\n" + pyl.Repository.Name + "\n" + pyl.Ref + "\n" + pyl.Repository.URL + "\n" + komsg
		dt := &model.TextMessage{
			To:       prj.Owner.PhoneNumber,
			IsGroup:  false,
			Messages: msg,
		}
		if prj.WAGroupID != "" {
			dt.To = prj.WAGroupID
			dt.IsGroup = true
		}
		resp, err = helper.PostStructWithToken[model.Response]("Token", config.WAAPIToken, dt, config.WAAPIMessage)
		if err != nil {
			resp.Info = "Tidak berhak"
			resp.Response = err.Error()
			helper.WriteJSON(respw, http.StatusUnauthorized, resp)
			return
		}
	}
	helper.WriteJSON(respw, http.StatusOK, resp)
}

func getMemberByAttributeInProject(project model.Project, attribute string, value string) (*model.Userdomyikado, error) {
	for _, member := range project.Members {
		switch attribute {
		case "email":
			if member.Email == value {
				return &member, nil
			}
		case "githubusername":
			if member.GithubUsername == value {
				return &member, nil
			}
		default:
			return nil, errors.New("unknown attribute")
		}
	}
	return nil, errors.New("member not found")
}
