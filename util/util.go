package util

import (
	"bufio"
	"fmt"
	"github.com/one-meta/meta-g/common"
	"go/token"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/Xuanwo/gg"
	mapset "github.com/deckarep/golang-set"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

var (
	importPkg = map[string]string{
		"context": "context",
		"driver":  "database/sql/driver",
		"errors":  "errors",
		"fmt":     "fmt",
		"math":    "math",
		"strings": "strings",
		"time":    "time",
		"ent":     "entgo.io/ent",
		"dialect": "entgo.io/ent/dialect",
		"schema":  "entgo.io/ent/schema",
		"field":   "entgo.io/ent/schema/field",
	}
	src = rand.NewSource(time.Now().UnixNano())
)

// Receiver returns the Receiver name of the given type.
//
//	[]T       => t
//	[1]T      => t
//	User      => u
//	UserQuery => uq
func Receiver(s string) (r string) {
	// Trim invalid tokens for identifier prefix.
	s = strings.Trim(s, "[]*&0123456789")
	parts := strings.Split(Snake(s), "_")
	min := len(parts[0])
	for _, w := range parts[1:] {
		if len(w) < min {
			min = len(w)
		}
	}
	for i := 1; i < min; i++ {
		r := parts[0][:i]
		for _, w := range parts[1:] {
			r += w[:i]
		}
		if _, ok := importPkg[r]; !ok {
			s = r
			break
		}
	}
	name := strings.ToLower(s)
	if token.Lookup(name).IsKeyword() {
		name = "_" + name
	}
	return name
}

// Snake converts the given struct or field name into a snake_case.
//
//	Username => username
//	FullName => full_name
//	HTTPCode => http_code
func Snake(s string) string {
	var (
		j int
		b strings.Builder
	)
	for i := 0; i < len(s); i++ {
		r := rune(s[i])
		// Put '_' if it is not a start or end of a word, current letter is uppercase,
		// and previous is lowercase (cases like: "UserInfo"), or next letter is also
		// a lowercase and previous letter is not "_".
		if i > 0 && i < len(s)-1 && unicode.IsUpper(r) {
			if unicode.IsLower(rune(s[i-1])) ||
				j != i-1 && unicode.IsLower(rune(s[i+1])) && unicode.IsLetter(rune(s[i-1])) {
				j = i
				b.WriteString("_")
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func goFmtFile(fileName string) {
	fmt.Printf("生成并格式化 %s\n", fileName)
	cmd := exec.Command("go", "fmt", fileName)
	err := cmd.Run()
	if err != nil {
		return
	}
}

func CheckGenFile(rootPath, target string, fileName string, generator *gg.Generator, merge ...bool) {
	targetName := filepath.Join(rootPath, target, fileName+".go")
	if merge != nil && merge[0] || common.Merge {
		err := generator.WriteFile(targetName)
		goFmtFile(targetName)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err := os.Stat(targetName)
		if err != nil {
			err = generator.WriteFile(targetName)
			goFmtFile(targetName)
		} else {
			fmt.Printf("即将跳过 %s\n", targetName)
		}
	}
}

func CheckDirAndMk(rootPath, dirName string) {
	_, err := os.Stat(rootPath + common.Separator + dirName)
	if err != nil {
		err := os.MkdirAll(rootPath+common.Separator+dirName, os.ModePerm)
		if err != nil {
			return
		}
	}
}

func GenWireSet(rootPath, packageName string) {
	generator := gg.New()
	group := generator.NewGroup()

	// 包名
	group.AddPackage(packageName)

	// import
	group.NewImport().AddPath("github.com/google/wire")

	setGroup := make([]string, 0, len(common.DataMap))
	for k := range common.DataMap {
		setGroup = append(setGroup, k+"Set")
	}
	group.AddLineComment("Controller wire set")
	newVar := group.NewVar().AddField("\nSet", fmt.Sprintf("wire.NewSet(%s)", strings.Join(setGroup, ",")))
	for _, v := range setGroup {
		controllerSet := strings.ReplaceAll(v, "Set", "Controller")
		newVar.AddField(v, fmt.Sprintf("wire.NewSet(wire.Struct(new(%s), \"*\"))", controllerSet))
	}

	// fmt.Println(group.String())
	CheckDirAndMk(rootPath, packageName)
	CheckGenFile(rootPath, packageName, "wire_set", generator)
}

// EntDataTransfer 提取data_map.go的数据
// dataMap data开头的数据，AssetIp = assetip
// fieldMap field开头的数据，AssetIp = StructField []
// defaultDataMap default开头的数据，User = Valid:true []
// timeSet time开头的数据，CreatedAt,UpdatedAt
func EntDataTransfer(fileName string) (map[string]string, map[string][]string, mapset.Set, map[string][]common.StructField) {
	dataMap := make(map[string]string)
	fieldMap := make(map[string][]common.StructField)
	defaultDataMap := make(map[string][]string)
	timeSet := mapset.NewSet()
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	line := bufio.NewReader(file)
	for {
		content, _, err := line.ReadLine()
		if err == io.EOF {
			break
		}
		str := string(content)
		str = replacePrefix(str, "//")
		if len(str) != 0 {
			if strings.HasPrefix(str, "data ") {
				str = replacePrefix(str, "data")
				split := strings.Split(str, "=")
				dataMap[split[0]] = split[1]
			}
			if strings.HasPrefix(str, "time ") {
				timeSet.Add(replacePrefix(str, "time"))
			}
			if strings.HasPrefix(str, "default ") {
				str = replacePrefix(str, "default")
				split := strings.Split(str, ",")
				key0 := split[0]
				key1 := split[1]

				// 存在
				if _, ok := defaultDataMap[key0]; ok {
					data := defaultDataMap[key0]
					data = append(data, key1)
					defaultDataMap[key0] = data
				} else {
					// 不存在
					var s1 []string
					s1 = append(s1, key1)
					defaultDataMap[key0] = s1
				}
			}

			// map structName , fieldList
			if strings.HasPrefix(str, "field ") {
				str = replacePrefix(str, "field")
				split := strings.Split(str, ",")
				nillable, _ := strconv.ParseBool(split[4])
				structField := &common.StructField{FiledName: split[1], FieldLength: split[2], FieldType: split[3], Nillable: nillable}
				structName := split[0]
				// 存在
				if _, ok := fieldMap[structName]; ok {
					data := fieldMap[structName]
					data = append(data, *structField)
					fieldMap[structName] = data
				} else {
					// 不存在
					var s1 []common.StructField
					s1 = append(s1, *structField)
					fieldMap[structName] = s1
				}
			}
		}
	}
	return dataMap, defaultDataMap, timeSet, fieldMap
}

func replacePrefix(str, prefix string) string {
	return strings.ReplaceAll(str, prefix+" ", "")
}

// RemoveZero 移除列表的零值
func RemoveZero(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if IfZero(v) {
			slice = append(slice[:i], slice[i+1:]...)
			return RemoveZero(slice)
		}
	}
	return slice
}

// IfZero 判断一个值是否为零值，只支持string,float,int,time 以及其各自的指针
func IfZero(arg interface{}) bool {
	if arg == nil {
		return true
	}
	switch v := arg.(type) {
	case int, int32, int16, int64:
		if v == 0 {
			return true
		}
	case float32:
		r := float64(v)
		return math.Abs(r-0) < 0.0000001
	case float64:
		return math.Abs(v-0) < 0.0000001
	case string:
		if v == "" {
			return true
		}
	case *string, *int, *int64, *int32, *int16, *int8, *float32, *float64, *time.Time:
		if v == nil {
			return true
		}
	case time.Time:
		return v.IsZero()
	default:
		return false
	}
	return false
}

func RandomValue(field common.StructField) (any, string) {
	now := time.Now()
	length, _ := strconv.Atoi(field.FieldLength)
	if length == 0 || length > 100 {
		if strings.Contains(field.FieldType, "int") || strings.Contains(field.FieldType, "float") {
			length = 3
		} else {
			length = 10
		}
	}
	switch field.FieldType {
	case "int", "int32", "int16", "int64":
		return fmt.Sprintf("%d%v", length, rand.New(rand.NewSource(now.UnixNano())).Int31n(100)), "int"
	case "float32":
		return fmt.Sprintf("%03v", rand.New(rand.NewSource(now.UnixNano())).Float32()), "int"
	case "float64":
		return fmt.Sprintf("%03v", rand.New(rand.NewSource(now.UnixNano())).Float64()), "int"
	case "string":
		return fmt.Sprintf("\"%s\"", randStr(length)), "string"
	// case *string, *int, *int64, *int32, *int16, *int8, *float32, *float64, *time.Time:
	case "time.Time":
		return "time.Now().AddDate(0, 0, -1)", ""
	}
	return nil, ""
}

func randStr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

func UpperCaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func LowerCaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func InitData(filePath string) {
	if filePath == "" {
		filePath = "data_map.go"
	}
	_, errPwd := os.Stat(".air.toml")
	_, errUp := os.Stat("../.air.toml")
	// 文件存在，Meta项目中执行
	if errPwd == nil || errUp == nil {
		filePath = filepath.Join("data", "data_map.go")
		fmt.Println("备份query_param.go")
		// app/ent/extend/query_param.go
		source, err := os.Open(fmt.Sprintf("app%sent%sextend%squery_param.go", common.Separator, common.Separator, common.Separator))
		if err != nil {
			log.Fatal(err)
		}
		defer source.Close()
		destination, err := os.Create(fmt.Sprintf("data%squery_param.go", common.Separator))
		if err != nil {
			log.Fatal(err)
		}
		defer destination.Close()
		_, err = io.Copy(destination, source)
		if err != nil {
			log.Println("备份失败")
		}

	} else {
		// 开发测试
		common.ApiPath = "app"
		common.ExtendPath = "app"
	}
	common.DataMap, common.DefaultDataMap, common.TimeSet, common.FieldMap = EntDataTransfer(filePath)
}
