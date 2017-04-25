package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type User struct {
	Username   string `json:"username"`
	Webring    bool   `json:"webring"`
	PublicHtml bool   `json:"-"`
	Quota      Quota  `json:"quota"`
}

type Quota struct {
	User  string      `json:"-"`
	Block QuotaLimits `json:"block_limits"`
	File  QuotaLimits `json:"file_limits"`
}

type QuotaLimits struct {
	Used  string `json:"used"`
	Soft  string `json:"soft"`
	Hard  string `json:"hard"`
	Grace string `json:"grace"`
}

func UserList(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("/home")
	if err != nil {
		marsh, _ := json.Marshal(ErrorResponse{err.Error()})
		w.Write(marsh)
		return
	}

	var users []User
	msgchan := make(chan User)
	num_responded := 0
	quotas, _ := getAllQuotas()

	for _, file := range files {
		go checkUser(file, msgchan)
	}

	for num_responded < len(files) {
		user := <-msgchan
		num_responded += 1
		user.Quota = quotas[user.Username]

		if user.PublicHtml {
			users = append(users, user)
		}
	}

	marsh, _ := json.Marshal(users)
	w.Write(marsh)
}

func checkUser(file os.FileInfo, ret chan User) {
	user := User{}
	if file.IsDir() {
		user.Username = file.Name()
		sub, _ := ioutil.ReadDir(fmt.Sprintf("/home/%s", user.Username))
		for _, subfile := range sub {
			if subfile.Name() == "public_html" {
				user.PublicHtml = true
			}
		}
	}
	user.Webring = isWebringMember(user.Username)
	ret <- user
}

func isWebringMember(user string) bool {
	sub, _ := ioutil.ReadDir(fmt.Sprintf("/home/%s", user))
	for _, subfile := range sub {
		if subfile.Name() == "webring" {
			return true
		}
	}

	return false
}

func getAllQuotas() (map[string]Quota, error) {
	quotaCmd := exec.Command("repquota", "-h", "-u", "/")
	result, err := quotaCmd.Output()
	if err != nil {
		return nil, err
	}

	rows := strings.Split(string(result), "\n")
	rows = rows[2 : len(rows)-3]
	ret := make(map[string]Quota)

	for _, row := range rows {
		fields := strings.Fields(row)
		q := Quota{
			User: fields[0],
			Block: QuotaLimits{
				Used:  fields[2],
				Soft:  fields[3],
				Hard:  fields[4],
				Grace: fields[5],
			},
			File: QuotaLimits{
				Used:  fields[6],
				Soft:  fields[7],
				Hard:  fields[8],
				Grace: fields[9],
			},
		}

		ret[q.User] = q
	}

	return ret, nil
}
