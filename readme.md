# Auto Selfcheck

## 설치
```bash
go get github.com/gangjun06/auto-selfcheck
```

## 사용법

[go Document]()

```go
// 학교 검색하기
orgCode, err := selfcheck.FindSchool("학교이름", Area, Level)
if err != nil{
    log.Fatal(err)
}

// 학생 정보 가져오기
info, err := selfcheck.GetStudentInfo(Area, orgCode, "이름", "생일(주민등록번호 앞자리)")
if err != nil{
    log.Fatal(err)
}

// 모두 건강함으로 참여하기
if err := info.AllHealthy(); err != nil {
    log.Fatal(err)
}

```

Area Type
```
AREA_SEOUL
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
```

Level Type
```
LEVEL_KIDER
LEVEL_ELEMENTRY
LEVEL_MIDDLE
LEVEL_HIGH
LEVEL_SPECIAL
)
```

## License
[Mit License](https://github.com/gangjun06/auto-selfcheck/blob/master/LICENSE)