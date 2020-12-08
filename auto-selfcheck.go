package selfcheck

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	netUrl "net/url"

	"github.com/imroc/req"
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
	Name       string `json:"userName"`
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

	r, _ := req.Get(url)

	var data schoolFind
	r.ToJSON(&data)

	if len(data.SchulList) < 1 {
		return "", ErrInfoNotFound
	}

	return data.SchulList[0].OrgCode, nil
}

// GetStudentInfo get student info struct
func GetStudnetInfo(area Area, orgCode, name, birth string) (*StudentInfo, error) {
	areaURL := GetAreaURL(area)
	url := fmt.Sprintf("https://%shcs.eduro.go.kr/v2/findUser", areaURL)
	reqBody := map[string]interface{}{
		"name":      Encrypt(name),
		"birthday":  Encrypt(birth),
		"orgCode":   orgCode,
		"loginType": "school",
	}

	r, _ := req.Post(url, req.BodyJSON(reqBody))

	var data StudentInfo
	if err := r.ToJSON(&data); err != nil {
		return nil, ErrInfoNotFound
	}

	url2 := fmt.Sprintf("https://%shcs.eduro.go.kr/v2/selectUserGroup", areaURL)
	header2 := req.Header{
		"Authorization": data.Token,
	}

	r2, _ := req.Post(url2, header2)

	var data2 []map[string]interface{}
	if err := r2.ToJSON(&data2); err != nil {
		return nil, err
	}

	userPNo := data2[0]["userPNo"].(string)
	token2 := data2[0]["token"].(string)

	url3 := fmt.Sprintf("https://%shcs.eduro.go.kr/v2/getUserInfo", areaURL)
	header3 := req.Header{
		"Authorization": token2,
	}
	reqBody3 := map[string]interface{}{
		"orgCode": orgCode,
		"userPNo": userPNo,
	}

	r3, _ := req.Post(url3, header3, req.BodyJSON(reqBody3))

	var data3 map[string]interface{}
	if err := r3.ToJSON(&data3); err != nil {
		return nil, err
	}

	data.Token = data3["token"].(string)
	data.AreaURL = areaURL
	data.Birth = birth

	return &data, nil
}

// AllHealthy Send Servey all healthy
func (s *StudentInfo) AllHealthy() error {
	url := fmt.Sprintf("https://%shcs.eduro.go.kr/registerServey", s.AreaURL)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"deviceUuid":         "",
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

	reqHeader := req.Header{
		"Authorization": s.Token,
		"Content-Type":  "application/json",
	}

	r, err := req.Post(url, reqHeader, req.BodyJSON(reqBody))
	if err != nil {
		return err
	}

	if r.Response().StatusCode != 200 {
		return errors.New("error request")
	}

	return nil
}
