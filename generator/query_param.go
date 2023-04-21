package generator

import (
	"fmt"
	"github.com/Xuanwo/gg"
	"github.com/one-meta/meta-g/common"
	"github.com/one-meta/meta-g/util"
)

// GenQueryParam 生成时间参数结构体
func GenQueryParam(rootPath string, merge bool) {
	generator := gg.New()
	group := generator.NewGroup()

	group.AddLineComment("Code generated by meta-generator, DO NOT EDIT.")
	packageName := "extend"
	//包名
	group.AddPackage(packageName)

	//import
	group.NewImport().AddPath("time")

	group.AddLineComment("时间参数结构体")
	newStruct := group.NewStruct("TimeParam")
	for v := range common.TimeSet.Iterator().C {
		//newStruct.AddField(fmt.Sprintf("\n%sGte", v), fmt.Sprintf("time.Time `query:\"%sGte\"`", v))
		//newStruct.AddField(fmt.Sprintf("%sLte", v), fmt.Sprintf("time.Time `query:\"%sLte\"`", v))

		newStruct.AddField(fmt.Sprintf("\n%sGte", v), fmt.Sprintf("time.Time `query:\"%s_gte\"`", util.Snake(v.(string))))
		newStruct.AddField(fmt.Sprintf("%sLte", v), fmt.Sprintf("time.Time `query:\"%s_lte\"`", util.Snake(v.(string))))
	}

	//fmt.Println(group.String())
	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, "query_param", generator, merge)
}
