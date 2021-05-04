package datahandling

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
)

var (
	statsURL       = "https://app.splatoon2.nintendo.net/api/records/hero"
	resultsURL     = "https://app.splatoon2.nintendo.net/api/results"
	nameAndIconURL = "https://app.splatoon2.nintendo.net/api/nickname_and_icon"
	useragent      = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_3 like Mac OS X) AppleWebKit/603.3.8 (KHTML, like Gecko) Mobile/14G60"
)

type Contestant struct {
	Name         string       `json:"name"`
	Cookie       string       `json:"cookie"`
	SessionToken string       `json:"session_token"`
	SplatnetName string       `json:"splatnet_name"`
	PictureURL   string       `json:"picture_url"`
	Date         time.Time    `json:"time"`
	SplatnetData SplatnetData `json:"splatnet_Data"`
}

//Stuff for getting picture and name

type BattleResults struct {
	BattleResults [50]BattleResult `json:"results,omitempty"`
}

type BattleResult struct {
	PlayerResult PlayerBattleResult `json:"player_result,omitempty"`
}

type PlayerBattleResult struct {
	Player Player `json:"player,omitempty"`
}

type Player struct {
	PrincipalID string `json:"principal_id,omitempty"`
}

type SplatnetProfiles struct {
	SplatnetProfiles []SplatnetProfile `json:"nickname_and_icons"`
}

type SplatnetProfile struct {
	Name       string `json:"nickname"`
	PictureURL string `json:"thumbnail_url"`
}

//Stuff for getting times
type SplatnetData struct {
	SplatnetCampaignSummary SplatnetCampaignSummary  `json:"summary"`
	SplatnetStageClearDatas []SplatnetStageClearData `json:"stage_infos"`
}

type SplatnetStageClearData struct {
	SplatnetStage SplatnetStage                      `json:"stage"`
	ClearWeapons  map[string]SplatnetWeaponClearData `json:"clear_weapons"`
}

type SplatnetWeaponClearData struct {
	ClearTime uint `json:"clear_time"`
}

type SplatnetStage struct {
	ID     string `json:"id"`
	IsBoss bool   `json:"is_boss"`
	Area   string `json:"area"`
}

type SplatnetCampaignSummary struct {
	SplatnetHonor SplatnetHonor `json:"honor"`
	ClearRate     float64       `json:"clear_rate"`
}

type SplatnetHonor struct {
	Name string `json:"name"`
}

func getContestant() (Contestant, error) {
	contestant, err := loadContestant()
	if err != nil {
		return contestant, err
	}
	err = loadSplatnetData(&contestant)
	if err != nil {
		return contestant, err
	}
	err = saveContestant(contestant)
	if err != nil {
		return contestant, err
	}
	return contestant, nil
}

func loadContestant() (contestant Contestant, err error) {
	byteValue, err := ioutil.ReadFile("contestant.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(byteValue, &contestant)
	return
}

func saveContestant(contestant Contestant) (err error) {
	output, err := json.MarshalIndent(contestant, "", "\t")
	if err != nil {
		return
	}
	err = ioutil.WriteFile("contestant.json", output, 0644)
	return
}

func loadSplatnetData(contestant *Contestant) error {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	//stats
	req, err := http.NewRequest("GET", statsURL, nil)
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("Cookie", "iksm_session="+contestant.Cookie)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(&contestant.SplatnetData)
	if err != nil {
		return err
	}
	//name and picture
	req, err = http.NewRequest("GET", resultsURL, nil)
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("Cookie", "iksm_session="+contestant.Cookie)
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	var results BattleResults
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return err
	}

	req, err = http.NewRequest("GET", nameAndIconURL, nil)
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("Cookie", "iksm_session="+contestant.Cookie)
	q := req.URL.Query()
	q.Set("id", results.BattleResults[0].PlayerResult.Player.PrincipalID)
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	var profiles SplatnetProfiles
	err = json.NewDecoder(resp.Body).Decode(&profiles)
	if err != nil {
		return err
	}
	if len(profiles.SplatnetProfiles) == 0 {
		fmt.Printf("Invalid cookie for %s, attempting to generate new.\n", contestant.Name)
		newCookie, err := generateCookie(contestant.SessionToken)
		if err != nil {
			return err
		}
		contestant.Cookie = newCookie
		err = saveContestant(*contestant)
		if err != nil {
			return err
		}
		return loadSplatnetData(contestant)
	}
	contestant.SplatnetName = profiles.SplatnetProfiles[0].Name
	contestant.PictureURL = profiles.SplatnetProfiles[0].PictureURL
	contestant.Date = time.Now()

	return nil
}

func generateCookie(sessionToken string) (string, error) {
	cmd := exec.Command("python3.9", "datahandling/iksm.py", sessionToken)
	out, err := cmd.Output()
	return string(out), err
}
