package main

import (
	"a21hc3NpZ25tZW50/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func GetStudyProgram() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Cek apakah method yang digunakan adalah GET
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Method is not allowed!"})
			return
		}
		// Membaca file list-study.txt
		studyData, err := ioutil.ReadFile("data/list-study.txt")
		if err != nil {
			// Jika gagal membaca, kirim HTTP status code 500 dan pesan error
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Internal Server Error"})
			return
		}
		// Buat slice untuk menampung data program studi
		var studies []model.StudyData
		lines := strings.Split(string(studyData), "\n")
		// Looping untuk memisahkan kode program studi dan namanya
		for _, line := range lines {
			fields := strings.Split(line, "_")
			if len(fields) == 2 {
				study := model.StudyData{
					Code: fields[0],
					Name: fields[1],
				}
				studies = append(studies, study)
			}
		}
		// Kirim HTTP status code 200 dan data program studi dalam format JSON
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(studies)
	}
}

func AddUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Memeriksa apakah metodenya adalah POST
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Method is not allowed!"})
			return
		}
		// Membaca body request dan memasukkannya ke variabel user
		var user model.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Body request tidak valid"})
			return
		}

		// Memeriksa apakah ID, nama, atau study kosong
		if user.ID == "" || user.Name == "" || user.StudyCode == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "ID, name, or study code is empty"})
			return
		}

		// Memeriksa apakah ID user sudah ada di file
		users, err := ioutil.ReadFile("data/list-study.txt")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Internal server error"})
			return
		}
		lines := strings.Split(string(users), "\n")
		for _, line := range lines {
			fields := strings.Split(line, "_")
			if len(fields) == 2 && fields[0] == user.StudyCode {
				if strings.Contains(line, user.ID) {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(model.ErrorResponse{Error: "user id already exist"})
					return
				}
			}
		}

		// Memeriksa apakah kode study sudah ada di file
		studyData, err := ioutil.ReadFile("data/list-study.txt")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Internal server error"})
			return
		}
		var studyExist bool
		for _, line := range strings.Split(string(studyData), "\n") {
			if strings.HasPrefix(line, user.StudyCode+"_") {
				studyExist = true
				break
			}
		}
		if !studyExist {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "study code not found"})
			return
		}

		// Menambahkan user ke file
		f, err := os.OpenFile("data/users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		_, err = fmt.Fprintf(f, "%s,%s,%s\n", user.ID, user.Name, user.StudyCode)
		if err != nil {
			panic(err)
		}

		// Mengembalikan respons sukses
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model.SuccessResponse{
			Username: user.ID,
			Message:  "add user success",
		})
	}
}

func DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Cek apakah method yang digunakan adalah DELETE
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Method is not allowed!"})
			return
		}

		// Ambil ID user dari parameter query URL
		userID := r.URL.Query().Get("id")
		if userID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "user id is empty"})
			return
		}

		// Baca data list user
		usersData, err := ioutil.ReadFile("data/users.txt")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Cari user yang akan dihapus
		lines := strings.Split(string(usersData), "\n")
		var found bool
		var newLines []string
		for _, line := range lines {
			fields := strings.Split(line, ",")
			if len(fields) == 3 && fields[0] == userID {
				found = true
				continue
			}
			if line != "" {
				newLines = append(newLines, line)
			}
		}
		if !found {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "user id not found"})
			return
		}

		// Tulis list user yang telah diperbarui ke file
		newData := []byte(strings.Join(newLines, "\n"))
		err = ioutil.WriteFile("data/users.txt", newData, 0644)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(model.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Mengembalikan respons yang sukses
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model.SuccessResponse{
			Username: userID,
			Message:  "delete success",
		})
	}
}
func main() {
	http.HandleFunc("/study-program", GetStudyProgram())
	http.HandleFunc("/user/add", AddUser())
	http.HandleFunc("/user/delete", DeleteUser())

	fmt.Println("starting web server at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
