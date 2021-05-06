package datahandling

import (
	"encoding/json"
	"net/http"
	"time"
)

func MaybeFetchAndSendData(w http.ResponseWriter) error {
	contestant, err := loadContestant()
	if err != nil || time.Now().Sub(contestant.Date) > 10*time.Minute {
		for attempts := 0; attempts < 5; attempts++ {
			contestant, err = getContestant()
			if err == nil {
				break
			}
			time.Sleep(15 * time.Second)
		}
	}
	contestant.Cookie = ""
	contestant.SessionToken = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contestant)
	return err
}
