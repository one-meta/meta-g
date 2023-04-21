package generator

import (
	"fmt"
	"github.com/Xuanwo/gg"
	"github.com/one-meta/meta-g/common"
	"github.com/one-meta/meta-g/util"
)

// GenApiPublic 生成公有路由代码
func GenApiPublic(rootPath string) {
	routerName := "Router"
	receiverName := util.Receiver(routerName)
	paraName := util.Receiver("FiberApp")

	generator := gg.New()
	group := generator.NewGroup()

	packageName := common.ApiVersion
	// 包名
	group.AddPackage(packageName)

	// import
	group.NewImport().
		AddPath("github.com/gofiber/fiber/v2")

	// Public 公共路由
	group.AddLineComment("Public 公共路由")
	newFunction := group.NewFunction("Public")
	newFunction.
		// 接收
		WithReceiver(receiverName, "*"+routerName).
		// 参数
		AddParameter(paraName, "*fiber.App").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"api := %s.Group(\"/%s\")\n"+
			"%s := api.Group(\"/%s\")",
			paraName, common.ApiName, packageName, packageName))
	for k, v := range common.DataMap {
		pkgRouterName := util.LowerCaseFirst(k) + routerName

		// 用户
		if k == "User" {
			newFunction.AddBody("//" + k + "路由\n" +
				fmt.Sprintf("%s := %s.Group(\"%s\")\n", pkgRouterName, packageName, v) +
				fmt.Sprintf("%s.Get(\"logout\", %s.%sController.Logout)\n", pkgRouterName, receiverName, k) +
				fmt.Sprintf("%s.Post(\"login\", %s.%sController.Login)\n", pkgRouterName, receiverName, k) +
				"\n",
			)
		}
	}
	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, "public", generator)
}

// GenApiPrivate 生成私有路由代码
func GenApiPrivate(rootPath string) {
	routerName := "Router"
	receiverName := util.Receiver(routerName)
	paraName := util.Receiver("FiberApp")

	generator := gg.New()
	group := generator.NewGroup()

	packageName := common.ApiVersion
	// 包名
	group.AddPackage(packageName)

	group.NewImport().
		AddPath("github.com/gofiber/fiber/v2").
		AddPath("github.com/one-meta/meta/app/entity/config")

	// import

	// Private 私有路由
	group.AddLineComment("Private 私有路由")
	newFunction := group.NewFunction("Private")
	newFunction.
		// 接收
		WithReceiver(receiverName, "*"+routerName).
		// 参数
		AddParameter(paraName, "*fiber.App").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"var router fiber.Router\n"+
			"//dev环境未启用认证授权\n"+
			"if !config.CFG.Auth.Enable && config.CFG.Stage.Status == \"dev\" {\n"+
			"router = %s.Group(\"/%s\", r.Gen.SaJWT(), r.Gen.SaCtx(), r.Gen.SaCasbin())\n"+
			"} else {\n"+
			"router = %s.Group(\"/%s\", %s.JWT.AuthJWT(), %s.Casbinx.AuthCasbin(%s.Enf), r.Gen.AuthCtx())\n"+
			"}\n"+
			"%s := router.Group(\"/%s\")\n",
			paraName, common.ApiName, paraName, common.ApiName, receiverName, receiverName, receiverName, packageName, packageName))
	for k, v := range common.DataMap {
		pkgRouterName := util.LowerCaseFirst(k) + routerName
		//casbinrule
		//if k == "CasbinRule" {
		//	newFunction.AddBody("//" + k + "路由\n" +
		//		fmt.Sprintf("%s := %s.Group(\"%s\")\n", pkgRouterName, packageName, v) +
		//		fmt.Sprintf("%s.Get(\"\", %s.%sController.Query)\n", pkgRouterName, receiverName, k) +
		//		fmt.Sprintf("%s.Get(\":id\", r.Gen.IntId(), %s.%sController.QueryByID)\n", pkgRouterName, receiverName, k) +
		//		fmt.Sprintf("%s.Post(\"\", r.Gen.CasbinRule(), %s.%sController.Create)\n", pkgRouterName, receiverName, k) +
		//		fmt.Sprintf("%s.Post(\"bulk\", %s.%sController.CreateBulk)\n", pkgRouterName, receiverName, k) +
		//		fmt.Sprintf("%s.Post(\"bulk/delete\", %s.%sController.DeleteBulk)\n", pkgRouterName, receiverName, k) +
		//		fmt.Sprintf("%s.Put(\":id\", r.Gen.IntId(), r.Gen.CasbinRule(), %s.%sController.UpdateByID)\n", pkgRouterName, receiverName, k) +
		//		fmt.Sprintf("%s.Delete(\":id\", r.Gen.IntId(), r.Gen.CasbinRule(), %s.%sController.DeleteByID)", pkgRouterName, receiverName, k) +
		//		"\n",
		//	)
		//} else {
		//
		//}

		newFunction.AddBody("//" + k + "路由\n" +
			fmt.Sprintf("%s := %s.Group(\"%s\")", pkgRouterName, packageName, v))
		if k == "User" {
			newFunction.AddBody(fmt.Sprintf("%s.Get(\"info\", %s.%sController.UserInfo)", pkgRouterName, receiverName, k))
		}
		newFunction.AddBody(
			fmt.Sprintf("%s.Get(\"\", %s.%sController.Query)\n", pkgRouterName, receiverName, k) +
				fmt.Sprintf("%s.Get(\":id\", r.Gen.IntId(), %s.%sController.QueryByID)\n", pkgRouterName, receiverName, k) +
				fmt.Sprintf("%s.Post(\"\", %s.%sController.Create)\n", pkgRouterName, receiverName, k) +
				fmt.Sprintf("%s.Post(\"bulk\", %s.%sController.CreateBulk)\n", pkgRouterName, receiverName, k) +
				fmt.Sprintf("%s.Post(\"bulk/delete\", %s.%sController.DeleteBulk)\n", pkgRouterName, receiverName, k) +
				fmt.Sprintf("%s.Put(\":id\", r.Gen.IntId(), %s.%sController.UpdateByID)\n", pkgRouterName, receiverName, k) +
				fmt.Sprintf("%s.Delete(\":id\", r.Gen.IntId(), %s.%sController.DeleteByID)", pkgRouterName, receiverName, k) +
				"\n",
		)

	}

	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, "private", generator)
}

// GenApiPrivateWireSet 生成路由wire set
func GenApiPrivateWireSet(rootPath string) {
	generator := gg.New()
	group := generator.NewGroup()
	structName := "Router"
	packageName := common.ApiVersion
	// 包名
	group.AddPackage(packageName)

	// import
	group.NewImport().
		AddPath("github.com/one-meta/meta/app/controller").
		AddPath("github.com/casbin/casbin/v2").
		AddPath("github.com/one-meta/meta/pkg/middleware").
		AddPath("github.com/google/wire").
		AddPath("github.com/one-meta/meta/pkg/middleware/generator")

	group.NewVar().AddField("Set", fmt.Sprintf("wire.NewSet(wire.Struct(new(%s), \"*\"))", structName))

	newStruct := group.NewStruct(structName)
	newStruct.
		AddField("Enf", "*casbin.Enforcer").
		AddField("Casbinx", "*middleware.Casbinx").
		AddField("JWT", "*middleware.JWT").
		AddField("Gen", "*generator.Gen")
	for k := range common.DataMap {
		newStruct.AddField(k+"Controller", fmt.Sprintf("*controller.%sController", k))
	}
	// fmt.Println(group.String())
	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, "wire_set", generator)
}

// GenApiNotFound 生成404路由代码
func GenApiNotFound(rootPath string) {
	routerName := "Router"
	receiverName := util.Receiver(routerName)
	paraName := util.Receiver("FiberApp")

	generator := gg.New()
	group := generator.NewGroup()

	packageName := common.ApiVersion
	// 包名
	group.AddPackage(packageName)

	// import
	group.NewImport().
		AddPath("github.com/gofiber/fiber/v2").
		AddPath("github.com/one-meta/meta/pkg/common")

	// NotFound 404路由
	group.AddLineComment("NotFound 404路由")
	newFunction := group.NewFunction("NotFound")
	newFunction.
		// 接收
		WithReceiver(receiverName, "*"+routerName).
		// 参数
		AddParameter(paraName, "*fiber.App").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"%s.Use(\n"+
			"func(c *fiber.Ctx) error {\n"+
			"return common.NewErrorWithStatusCode(c, \"Not Found\", fiber.StatusNotFound)\n"+
			"},\n"+
			")",
			paraName))

	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, "not_found", generator)
}

// GenApiSwagger 生成Swagger api路由代码
func GenApiSwagger(rootPath string) {
	routerName := "Router"
	receiverName := util.Receiver(routerName)
	paraName := util.Receiver("FiberApp")

	generator := gg.New()
	group := generator.NewGroup()

	packageName := common.ApiVersion
	// 包名
	group.AddPackage(packageName)

	// import
	group.NewImport().
		AddPath("github.com/gofiber/fiber/v2").
		AddPath("github.com/gofiber/swagger")

	// Swagger api路由
	group.AddLineComment("Swagger api路由")
	newFunction := group.NewFunction("Swagger")
	newFunction.
		// 接收
		WithReceiver(receiverName, "*"+routerName).
		// 参数
		AddParameter(paraName, "*fiber.App").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"api := %s.Group(\"/swagger\")\n"+
			"api.Get(\"*\", swagger.HandlerDefault)",
			paraName))

	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, "swagger", generator)
}
