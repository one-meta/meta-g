package main

import (
	"flag"
	"fmt"
	"github.com/one-meta/meta-g/common"
	"github.com/one-meta/meta-g/generator"
	"github.com/one-meta/meta-g/util"
	"os"
	"time"
)

var path string

func main() {
	flag.BoolVar(&common.Merge, "merge", false, "覆盖已生成文件")
	flag.StringVar(&path, "path", "", "data_map.go文件绝对路径")
	flag.Parse()
	if common.Merge {
		fmt.Println("覆盖已生成文件，操作将无法回滚，是否继续？[y/n]")
		var checkYes string
		fmt.Scan(&checkYes)
		if checkYes != "y" {
			os.Exit(0)
		}
	}
	fmt.Println("starting")
	now := time.Now()
	// 初始化数据

	flag.Parse()
	util.InitData(path)

	// 自动化测试api
	index := 10
	for entity, fields := range common.FieldMap {
		// randomIndex := rand.New(rand.NewSource(now.UnixNano())).Intn(100)
		generator.GenTestApi(entity, fields, index)
		index = index + 1
	}

	// 生成查询时间范围的参数
	generator.GenQueryParam(common.ExtendPath, true)

	// 生成service、controller
	for k, v := range common.DataMap {
		generator.GenService(k, common.ServicePath)
		generator.GenController(k, v, common.ControllerPath)
	}
	// 生成Service wire set
	generator.GenServiceWireSet(common.ServicePath)
	// 生成Controller wire set
	util.GenWireSet(common.ControllerPath, "controller")

	// 生成api
	// 生成私有路由及wire set
	generator.GenApiPrivate(common.ApiPath)
	generator.GenApiPrivateWireSet(common.ApiPath)
	// 生成404、公有、swagger api路由
	generator.GenApiNotFound(common.ApiPath)
	generator.GenApiPublic(common.ApiPath)
	generator.GenApiSwagger(common.ApiPath)
	fmt.Println("done.")
	fmt.Printf("耗时: %v\n", time.Since(now))
}
