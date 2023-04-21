package generator

import (
	"fmt"
	"github.com/one-meta/meta-g/common"
	"github.com/one-meta/meta-g/util"
	"strings"

	"github.com/Xuanwo/gg"
)

func GenTestApi(entity string, fields []common.StructField, index int) {
	rootPath := common.ApiName

	generator := gg.New()
	group := generator.NewGroup()

	packageName := common.ApiTest
	// 包名
	group.AddPackage(packageName)

	lowerCaseFirst := util.LowerCaseFirst(entity)
	lowerCaseEntity := strings.ToLower(entity)
	newEntity := "test" + entity
	idsName := lowerCaseFirst + "IDs"
	apiName := lowerCaseFirst + "Api"

	// import
	groupImport := group.NewImport().
		AddPath("github.com/one-meta/meta/app/ent").
		AddPath("github.com/one-meta/meta/app/ent/" + lowerCaseEntity).
		AddPath("github.com/one-meta/meta/pkg/common").
		AddPath("strconv").
		AddPath("testing")

	group.NewVar().AddField(apiName, fmt.Sprintf("baseApi + \"/%s\"", lowerCaseEntity))

	// lowerEntity := lowerCaseEntity
	dataSlice := make([]string, 10)
	valueSlice := make([]string, 10)
	var queryKey string
	// 判断是否有字符字段，用于查询
	stringFlag := false
	timeFlag := true
	for _, field := range fields {
		if field.FieldType == "string" {
			stringFlag = true
		}
		if strings.Contains(field.FieldType, "time.Time") {
			if timeFlag {
				groupImport.AddPath("time")
				timeFlag = false
			}
		}
	}
	valueSlice, dataSlice, queryKey = getRandomData(fields, stringFlag, index)
	dataJoin := strings.Join(dataSlice, ",")

	// var变量
	for _, v := range valueSlice {
		split := strings.Split(v, ":=")
		group.NewVar().AddField(split[0], split[1])
	}

	group.NewVar().AddField(newEntity, fmt.Sprintf("&ent.%s{%s}", entity, dataJoin))
	group.NewVar().AddField(idsName, "[]int{}")

	// 创建
	group.AddLineComment(fmt.Sprintf("%s %s test case", common.Create, entity))
	group.AddLineComment(common.CreateComment)
	createFunction := group.NewFunction(fmt.Sprintf("Test%s%s", common.Create, entity))
	createFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"successExpectedResult.ResponseData = %s.%s\n"+
			"testCase := &CaseRule{Api: %s, HttpMethod: \"POST\", BodyData: %s, Expected: successExpectedResult, Assert: assertEqualContains}\n"+
			"result := runTest(t, testCase)\n"+
			"%s = append(%s, getDataMapId(result))",
			newEntity, queryKey, apiName, newEntity, idsName, idsName))

	// 批量创建
	// 生成批量创建数据
	var bulkData []string
	for i := 1; i <= 5; i++ {
		valueSlice, dataSlice, _ := getRandomData(fields, stringFlag, i)
		if len(valueSlice) != 0 {
			valueJoin := strings.Join(valueSlice, "\n")
			bulkData = append(bulkData, valueJoin)
		}
		dataJoin := strings.Join(dataSlice, ",")
		bulkData = append(bulkData, fmt.Sprintf("bulkData%d := &ent.%s{%s}", i, entity, dataJoin))
	}
	noZeroBulkData := util.RemoveZero(bulkData)

	group.AddLineComment(fmt.Sprintf("%s %s test case", common.CreateBulk, entity))
	group.AddLineComment(common.CreateBulkComment)
	createBulkFunction := group.NewFunction(fmt.Sprintf("Test%s%s", common.CreateBulk, entity))
	createBulkFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"%s\n"+
			"bulkDatas := [...]ent.%s{*bulkData1, *bulkData2, *bulkData3, *bulkData4, *bulkData5}\n"+
			"successExpectedResult.ResponseData = bulkData1.%s\n"+
			"testCase := &CaseRule{Api: %s + \"/bulk\", HttpMethod: \"POST\", BodyData: bulkDatas, Expected: successExpectedResult, Assert: assertEqualContains}\n"+
			"result := runTest(t, testCase)\n\tdataMap := result.([]any)\n"+
			"for _, v := range dataMap {\n"+
			"%s = append(%s, getDataMapId(v))\n"+
			"}",
			strings.Join(noZeroBulkData, "\n"), entity, queryKey, apiName, idsName, idsName))

	// 查询
	group.AddLineComment(fmt.Sprintf("%s %s test case", common.Query, entity))
	group.AddLineComment(common.QueryComment)
	queryFunction := group.NewFunction(fmt.Sprintf("Test%s%s", common.Query, entity))
	queryFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"successExpectedResult.ResponseData = %s.%s\n"+
			"testCase := &CaseRule{Api: %s, HttpMethod: \"GET\", UrlData: \"current=1&pageSize=10\", Expected: successExpectedResult, Assert: assertEqualContainsGrater}\n"+
			"runTest(t, testCase)",
			newEntity, queryKey, apiName))

	// 根据ID查询
	group.AddLineComment(fmt.Sprintf("%s %s test case", common.QueryByID, entity))
	group.AddLineComment(common.QueryByIDComment)
	queryByIDFunction := group.NewFunction(fmt.Sprintf("Test%s%s", common.QueryByID, entity))
	queryByIDFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"successExpectedResult.ResponseData = %s.%s\n"+
			"testCase := &CaseRule{Api: %s, HttpMethod: \"GET\", UrlData: strconv.Itoa(%s[0]), Expected: successExpectedResult, Assert: assertEqualContains}\n"+
			"runTest(t, testCase)",
			newEntity, queryKey, apiName, idsName))

	// 根据ID查询不存在的
	group.AddLineComment(fmt.Sprintf("%s %s not exist test case", common.QueryByID, entity))
	group.AddLineComment(common.QueryByIDComment)
	queryByIDNotExistFunction := group.NewFunction(fmt.Sprintf("Test%s%sNotExist", common.QueryByID, entity))
	queryByIDNotExistFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"successExpectedResult.ResponseData = %s.%s\n"+
			"testCase := &CaseRule{Api: %s, HttpMethod: \"GET\", UrlData: \"0\", Expected: errorFoundExpectedResult, Assert: assertEqualContains}\n"+
			"runTest(t, testCase)",
			newEntity, queryKey, apiName))

	// 根据某个字段查询
	group.AddLineComment(fmt.Sprintf("%s %s by %s test case", common.Query, entity, queryKey))
	group.AddLineComment(common.QueryComment)
	queryByFieldFunction := group.NewFunction(fmt.Sprintf("Test%s%sBy%s", common.Query, entity, queryKey))
	queryByFieldFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"successExpectedResult.ResponseData = %s.%s\n"+
			"testCase := &CaseRule{Api: %s, HttpMethod: \"GET\", UrlData:  %s.Field%s + \"=\" + %s.%s, Expected: successExpectedResult, Assert: assertEqualContains}\n"+
			"runTest(t, testCase)",
			newEntity, queryKey, apiName, lowerCaseEntity, queryKey, newEntity, queryKey))

	// 根据某个字段搜索
	group.AddLineComment(fmt.Sprintf("%s %s search by %s test case", common.QuerySearch, entity, queryKey))
	group.AddLineComment(common.QuerySearchComment)
	querySearchFieldFunction := group.NewFunction(fmt.Sprintf("Test%s%s%s", common.QuerySearch, entity, queryKey))
	querySearchFieldFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"successExpectedResult.ResponseData = %s.%s\n"+
			"testCase := &CaseRule{Api: %s, HttpMethod: \"GET\", UrlData: \"search=\" + %s.%s, Expected: successExpectedResult, Assert: assertEqualContains}\n"+
			"runTest(t, testCase)",
			newEntity, queryKey, apiName, newEntity, queryKey))

	// 更新
	// 创建一条数据
	var dataUpdate []string
	valueSlice, dataSlice, queryKey = getRandomData(fields, stringFlag, index)
	if len(valueSlice) != 0 {
		valueJoin := strings.Join(valueSlice, "\n")
		dataUpdate = append(dataUpdate, valueJoin)
	}
	dataUpdate = append(dataUpdate, fmt.Sprintf("updateData := &ent.%s{%s}", entity, strings.Join(dataSlice, ",")))
	dataUpdate = append(dataUpdate)

	group.AddLineComment(fmt.Sprintf("%s %s test case", common.UpdateByID, entity))
	group.AddLineComment(common.UpdateByIDComment)
	updateFunction := group.NewFunction(fmt.Sprintf("Test%s%s", common.UpdateByID, entity))
	updateFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"%s\n"+
			"successExpectedResult.ResponseData = updateData.%s\n"+
			"testCase := &CaseRule{Api: %s, HttpMethod: \"PUT\", UrlData: strconv.Itoa(%s[0]), BodyData: updateData, Expected: successExpectedResult, Assert: assertEqualContains}\n"+
			"runTest(t, testCase)",
			strings.Join(dataUpdate, "\n"), queryKey, apiName, idsName))

	// 根据ID删除
	group.AddLineComment(fmt.Sprintf("%s %s test case", common.DeleteByID, entity))
	group.AddLineComment(common.DeleteByIDComment)
	deleteByIDFunction := group.NewFunction(fmt.Sprintf("Test%s%s", common.DeleteByID, entity))
	deleteByIDFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"testCase := &CaseRule{Api: %s, HttpMethod: \"DELETE\", UrlData: strconv.Itoa(%s[0]), Expected: successExpectedResult, Assert: assertEqual}\n"+
			"runTest(t, testCase)",
			apiName, idsName))

	// 根据ID删除不存在的
	group.AddLineComment(fmt.Sprintf("%s %s not exist test case", common.DeleteByID, entity))
	group.AddLineComment(common.DeleteByIDComment)
	deleteByIDNotExistFunction := group.NewFunction(fmt.Sprintf("Test%s%sNoExist", common.DeleteByID, entity))
	deleteByIDNotExistFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"\ttestCase := &CaseRule{Api: %s, HttpMethod: \"DELETE\", UrlData: \"0\", Expected: errorFoundExpectedResult, Assert: assertEqual}\n"+
			"runTest(t, testCase)", apiName))

	// 批量删除
	group.AddLineComment(fmt.Sprintf("%s %s test case", common.DeleteBulk, entity))
	group.AddLineComment(common.DeleteBulkComment)
	deleteBulkFunction := group.NewFunction(fmt.Sprintf("Test%s%s", common.DeleteBulk, entity))
	deleteBulkFunction.
		// 参数
		AddParameter("t", "*testing.T").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"testCase := &CaseRule{Api: %s + \"/bulk/delete\", HttpMethod: \"POST\", BodyData: common.DeleteItem{Ids: %s}, Expected: successExpectedResult, Assert: assertEqual}\n"+
			"runTest(t, testCase)",
			apiName, idsName))

	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, util.Snake(entity)+"_test", generator)
}

func getRandomData(fields []common.StructField, stringFlag bool, pointerIndex int) ([]string, []string, string) {
	var (
		valueSlice []string
		dataSlice  []string
		queryKey   string
	)

	for _, field := range fields {
		filedName := field.FiledName
		data, dataValue := util.RandomValue(field)
		// ID参数不设置
		if data != nil && !strings.Contains(filedName, "ID") {
			// 可为空的数据，需要传指针
			if field.Nillable {
				if pointerIndex != 0 {
					valueSlice = append(valueSlice, fmt.Sprintf("%s%d := %s", util.Snake(filedName), pointerIndex, data))
					dataSlice = append(dataSlice, fmt.Sprintf("%s: &%s%d", filedName, util.Snake(filedName), pointerIndex))
				} else {
					valueSlice = append(valueSlice, fmt.Sprintf("%s := %s", util.Snake(filedName), data))
					dataSlice = append(dataSlice, fmt.Sprintf("%s: &%s", filedName, util.Snake(filedName)))
				}
			} else {
				dataSlice = append(dataSlice, fmt.Sprintf("%s: %v", filedName, data))
				if stringFlag {
					if queryKey == "" && dataValue == "string" {
						queryKey = filedName
					}
				} else {
					if queryKey == "" && dataValue != "" {
						queryKey = filedName
					}
				}
			}
			// dataSlice = append(dataSlice, fmt.Sprintf("%s: %v", filedName, data))
		}
	}
	return util.RemoveZero(valueSlice), util.RemoveZero(dataSlice), queryKey
}
