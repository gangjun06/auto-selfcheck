package selfcheck

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	netUrl "net/url"
)

var (
	ErrInfoNotFound = errors.New("Cannot Find Data")
)

const (
	AREA_SEOUL = 1 + iota
	AREA_BUSAN
	AREA_DAEGU
	AREA_INCHEON
	AREA_GWANGJU
	AREA_DAEJEON
	AREA_ULSAN
	AREA_SEJONG
	AREA_GYEONGGI
	AREA_GANGWON
	AREA_CHUNGBUK
	AREA_CHUNGNAM
	AREA_JEONBUK
	AREA_JEONNAM
	AREA_GYEONGBUK
	AREA_GYEONGNAM
	AREA_JEJ
)

type Area int

const (
	LEVEL_KIDER = 1 + iota
	LEVEL_ELEMENTRY
	LEVEL_MIDDLE
	LEVEL_HIGH
	LEVEL_SPECIAL
)

type Level int

type schoolFind struct {
	SchulList []schulList `json:"schulList"`
}

type schulList struct {
	OrgCode string `json:"orgCode"`
}

type StudentInfo struct {
	SchoolName string `json:"orgname"`
	Name       string `json:"name"`
	Token      string `json:"token"`
	Birth      string
	AreaURL    string
}

// GetAreaCode from area code
func GetAreaCode(area Area) int {
	AreaCode := []int{1, 2, 3, 4, 5, 6, 7, 8, 10, 11, 12, 13, 14, 15, 16, 17, 18}
	return AreaCode[area-1]
}

// GetAreaURL for reqest to hcs.eduro.go.kr
func GetAreaURL(area Area) string {
	AreaURL := []string{"sen", "pen", "dge", "ice", "gen", "dje", "use", "sje", "goe", "kwe", "cbe", "cne", "jbe", "jne", "gbe", "gne", "jje"}
	return AreaURL[area-1]
}

// Encrypt For SignIn
func Encrypt(text string) *string {
	keyOrigin := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA81dCnCKt0NVH7j5Oh2+SGgEU0aqi5u6sYXemouJWXOlZO3jqDsHYM1qfEjVvCOmeoMNFXYSXdNhflU7mjWP8jWUmkYIQ8o3FGqMzsMTNxr+bAp0cULWu9eYmycjJwWIxxB7vUwvpEUNicgW7v5nCwmF5HS33Hmn7yDzcfjfBs99K5xJEppHG0qc+q3YXxxPpwZNIRFn0Wtxt0Muh1U8avvWyw03uQ/wMBnzhwUC8T4G5NclLEWzOQExbQ4oDlZBv8BM/WxxuOyu0I8bDUDdutJOfREYRZBlazFHvRKNNQQD2qDfjRz484uFs7b5nykjaMB9k/EJAuHjJzGs9MMMWtQIDAQAB"

	key, err := base64.StdEncoding.DecodeString(keyOrigin)
	if err != nil {
		log.Fatal(err)
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(key)
	if err != nil {
		log.Fatal(err)
	}

	publicKey, isRSAPublicKey := publicKeyInterface.(*rsa.PublicKey)
	if !isRSAPublicKey {
		log.Fatal("It is not RSA Public Key")
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(text))
	if err != nil {
		log.Fatal(err)
	}

	result := base64.StdEncoding.EncodeToString(ciphertext)

	return &result
}

// FindSchool get school orgCode
func FindSchool(name string, area Area, level Level) (string, error) {
	var url string
	lctnScCode := GetAreaCode(area)
	schulCrseCode := netUrl.QueryEscape(name)
	if lctnScCode < 10 {
		url = fmt.Sprintf("https://hcs.eduro.go.kr/v2/searchSchool?lctnScCode=0%d&schulCrseScCode=%d&orgName=%s&loginType=school", lctnScCode, level, schulCrseCode)
	} else {
		url = fmt.Sprintf("https://hcs.eduro.go.kr/v2/searchSchool?lctnScCode=%d&schulCrseScCode=%d&orgName=%s&loginType=school", lctnScCode, level, schulCrseCode)
	}
	fmt.Println(url)
	resp, _ := http.Get(url)
	dataByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var data schoolFind
	err := json.Unmarshal(dataByte, &data)
	if err != nil {
		return "", err
	}

	if len(data.SchulList) < 1 {
		return "", ErrInfoNotFound
	}

	return data.SchulList[0].OrgCode, nil
}

// GetStudentInfo get student info struct
func GetStudnetInfo(area Area, orgCode, name, birth string) (*StudentInfo, error) {
	areaURL := GetAreaURL(area)
	url := fmt.Sprintf("https://%shcs.eduro.go.kr/v2/findUser", areaURL)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"name":      Encrypt(name),
		"birthday":  Encrypt(birth),
		"orgCode":   orgCode,
		"loginType": "school",
	})

	reqBodyBuff := bytes.NewBuffer(reqBody)

	resp, err := http.Post(url, "application/json", reqBodyBuff)
	if err != nil {
		return nil, err
	}

	dataByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var data StudentInfo
	if err := json.Unmarshal(dataByte, &data); err != nil {
		return nil, ErrInfoNotFound
	}

	data.AreaURL = areaURL
	data.Birth = birth
	return &data, nil
}

// AllHealthy Send Servey all healthy
func (s *StudentInfo) AllHealthy() error {
	url := fmt.Sprintf("https://%shcs.eduro.go.kr/registerServey", s.AreaURL)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"eviceUuid":          "",
		"rspns00":            "Y",
		"rspns01":            "1",
		"rspns02":            "1",
		"rspns03":            nil,
		"rspns04":            nil,
		"rspns05":            nil,
		"rspns06":            nil,
		"rspns07":            nil,
		"rspns08":            nil,
		"rspns09":            "0",
		"rspns10":            nil,
		"rspns11":            nil,
		"rspns12":            nil,
		"rspns13":            nil,
		"rspns14":            nil,
		"rspns15":            nil,
		"upperToken":         s.Token,
		"upperUserNameEncpt": s.Name,
	})

	reqBodyBuff := bytes.NewBuffer(reqBody)
	req, err := http.NewRequest("POST", url, reqBodyBuff)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", s.Token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	dataByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(dataByte, &data); err != nil {
		return err
	}

	return nil
}
