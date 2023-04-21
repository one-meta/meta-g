package generator

import (
	"fmt"
	"github.com/one-meta/meta-g/common"
	"github.com/one-meta/meta-g/util"
	"strings"

	"github.com/Xuanwo/gg"
)

// GenController 生成Controller代码
func GenController(entityName, pkgName, rootPath string) {
	controllerName := entityName + "Controller"
	serviceName := entityName + "Service"
	// setName := entityName + "Set"
	receiverName := util.Receiver(controllerName)
	paraName := util.Receiver(entityName)
	queryParamName := util.Receiver("queryParam")
	routerName := pkgName

	generator := gg.New()
	group := generator.NewGroup()

	packageName := "controller"
	// 包名
	group.AddPackage(packageName)

	// import
	groupImport := group.NewImport().AddPath("github.com/one-meta/meta/app/ent").AddPath("github.com/one-meta/meta/app/service").
		AddPath("github.com/one-meta/meta/pkg/common").
		AddPath("github.com/gofiber/fiber/v2").
		AddPath("go.uber.org/zap").
		//AddPath("github.com/redis/go-redis/v9").
		AddPath("context")

	groupStruct := group.NewStruct(controllerName)
	groupStruct.
		AddField(serviceName, "*service."+serviceName).
		AddField("Logger", "*zap.Logger")
	//AddField("Rdb", "*redis.Client")

	// 查询 搜索
	// 生成 swagger api @Param 注解
	fields := common.FieldMap[entityName]
	var typeValue []string
	for _, field := range fields {
		fieldType := field.FieldType
		filedName := field.FiledName

		if !strings.Contains(strings.ToLower(filedName), "id") {
			var swaggerType, swaggerFormat string
			switch fieldType {
			case "int", "int32", "int16", "int64", "uint", "uint32", "uint16", "uint64":
				swaggerType = "integer"
			case "float32", "float64":
				swaggerType = "number"
			case "bool":
				swaggerType = "bool"
			case "time.Time":
				swaggerType = "string"
				swaggerFormat = "date-time"
			case "string":
				swaggerType = "string"
				if strings.EqualFold(filedName, "ip") {
					swaggerFormat = "ipv4"
				}
				if strings.EqualFold(filedName, "ipv6") {
					swaggerFormat = "ipv6"
				}
				if strings.EqualFold(filedName, "password") {
					swaggerFormat = "password"
				}
			}
			if swaggerType != "" {
				snakeField := util.Snake(filedName)
				if swaggerFormat != "" {
					typeValue = append(typeValue, fmt.Sprintf("@Param %s query %s false \"%s\" Format(%s)", snakeField, swaggerType, snakeField, swaggerFormat))
				} else {
					typeValue = append(typeValue, fmt.Sprintf("@Param %s query %s false \"%s\"", snakeField, swaggerType, snakeField))
				}
			}
		}
	}
	// 搜索
	typeValue = append(typeValue, "@Param search query string false \"需要搜索的值，多个值英文逗号,分隔\"")
	// 分页
	typeValue = append(typeValue, fmt.Sprintf("@Param %s query integer false \"%s\"", "current", "当前页"))
	typeValue = append(typeValue, fmt.Sprintf("@Param %s query integer false \"%s\"", "pageSize", "分页大小"))
	// 字段排序
	typeValue = append(typeValue, fmt.Sprintf("@Param %s query string false \"%s\"", "order", "排序，默认id逆序(-id)"))

	group.AddLineComment(
		fmt.Sprintf(common.SwaggerCommentPrefix,
			common.Query+common.QueryComment+entityName,
			common.Query+common.QueryComment+entityName,
			common.Query+common.QueryComment+entityName,
			entityName,
			strings.Join(typeValue, "\n"),
			fmt.Sprintf("common.Result{data=[]ent.%s}", entityName),
			fmt.Sprintf("/%s/%s/%s [get]", common.ApiName, common.ApiVersion, routerName),
		))
	group.NewFunction(common.Query).
		// 接收
		WithReceiver(receiverName, "*"+controllerName).
		// 参数
		AddParameter("c", "*fiber.Ctx").
		// 返回
		AddResult("", "error").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"%s := &ent.%s{}\n"+
			"%s, err := common.QueryParser(c, %s)\n"+
			"if err != nil {\n"+
			"return common.NewResult(c, err)\n"+
			"}\n"+
			"ctx := c.Locals(\"ctx\").(context.Context)\n"+
			"count, result, err := %s.%s.Query(ctx, %s, %s)\n"+
			"return common.NewPageResult(c, err, count, result)",
			paraName, entityName, queryParamName, paraName,
			receiverName, serviceName, paraName, queryParamName))

	// 根据ID查询
	group.AddLineComment(
		fmt.Sprintf(common.SwaggerCommentPrefix,
			common.QueryByID+common.QueryByIDComment+entityName,
			common.QueryByID+common.QueryByIDComment+entityName,
			common.QueryByID+common.QueryByIDComment+entityName,
			entityName,
			fmt.Sprintf("@Param id path int true \"%s ID\"", entityName),
			fmt.Sprintf("common.Result{data=ent.%s}", entityName),
			fmt.Sprintf("/%s/%s/%s/{id} [get]", common.ApiName, common.ApiVersion, routerName),
		))
	group.NewFunction(common.QueryByID).
		// 接收
		WithReceiver(receiverName, "*"+controllerName).
		// 参数
		AddParameter("c", "*fiber.Ctx").
		// 返回
		AddResult("", "error").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"id := c.Locals(\"id\").(int)\n"+
			"ctx := c.Locals(\"ctx\").(context.Context)\n"+
			"result, err := %s.%s.QueryByID(ctx, id)\n"+
			"return common.NewResult(c, err, result)",
			receiverName, serviceName))

	defaultValues := common.DefaultDataMap[entityName]
	var defaultValue string
	if len(defaultValues) != 0 {
		defaultValue = strings.Join(defaultValues, ",")
	}
	// 创建
	group.AddLineComment(
		fmt.Sprintf(common.SwaggerCommentPrefix,
			common.Create+common.CreateComment+entityName,
			common.Create+common.CreateComment+entityName,
			common.Create+common.CreateComment+entityName,
			entityName,
			fmt.Sprintf("@Param %s body ent.%s true \"%s\"", pkgName, entityName, entityName),
			fmt.Sprintf("common.Result{data=ent.%s}", entityName),
			fmt.Sprintf("/%s/%s/%s [post]", common.ApiName, common.ApiVersion, routerName),
		))

	group.NewFunction(common.Create).
		// 接收
		WithReceiver(receiverName, "*"+controllerName).
		// 参数
		AddParameter("c", "*fiber.Ctx").
		// 返回
		AddResult("", "error").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"%s := &ent.%s{%s}\n"+
			"err := common.BodyParser(c, %s)\n"+
			"if err != nil {\n"+
			"return common.NewResult(c, err)\n"+
			"}\n"+
			"ctx := c.Locals(\"ctx\").(context.Context)\n"+
			"create, err := %s.%s.Create(ctx, %s)\n"+
			"return common.NewResult(c, err, create)",
			paraName, entityName, defaultValue, paraName, receiverName, serviceName, paraName))

	// 批量创建
	group.AddLineComment(
		fmt.Sprintf(common.SwaggerCommentPrefix,
			common.CreateBulk+common.CreateBulkComment+entityName,
			common.CreateBulk+common.CreateBulkComment+entityName,
			common.CreateBulk+common.CreateBulkComment+entityName,
			entityName,
			fmt.Sprintf("@Param %s body []ent.%s true \"%s\"", pkgName, entityName, entityName),
			fmt.Sprintf("common.Result{data=[]ent.%s}", entityName),
			fmt.Sprintf("/%s/%s/%s/bulk [post]", common.ApiName, common.ApiVersion, routerName),
		))
	group.NewFunction(common.CreateBulk).
		// 接收
		WithReceiver(receiverName, "*"+controllerName).
		// 参数
		AddParameter("c", "*fiber.Ctx").
		// 返回
		AddResult("", "error").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"%s := make([]*ent.%s, 10)\n"+
			"err := common.RequestBodyParser(c, &%s)\n"+
			"if err != nil {\n"+
			"return common.NewResult(c, err)\n"+
			"}\n"+
			"ctx := c.Locals(\"ctx\").(context.Context)\n"+
			"bulkData, err := %s.%s.CreateBulk(ctx, %s)\n"+
			"return common.NewResult(c, err,bulkData)",
			paraName, entityName, paraName, receiverName, serviceName, paraName))

	// 修改
	group.AddLineComment(
		fmt.Sprintf(common.SwaggerCommentPrefix,
			common.UpdateByID+common.UpdateByIDComment+entityName,
			common.UpdateByID+common.UpdateByIDComment+entityName,
			common.UpdateByID+common.UpdateByIDComment+entityName,
			entityName,
			fmt.Sprintf("@Param id path int true \"%s ID\"\n", entityName)+
				fmt.Sprintf("@Param %s body ent.%s true \"%s\"", pkgName, entityName, entityName),
			fmt.Sprintf("common.Result{data=ent.%s}", entityName),
			fmt.Sprintf("/%s/%s/%s/{id} [put]", common.ApiName, common.ApiVersion, routerName),
		))

	group.NewFunction(common.UpdateByID).
		// 接收
		WithReceiver(receiverName, "*"+controllerName).
		// 参数
		AddParameter("c", "*fiber.Ctx").
		// 返回
		AddResult("", "error").
		// 方法体
		// AddBody(fmt.Sprintf(""+
		// 	"%s := &ent.%s{}\n"+
		// 	"err := common.BodyParser(c, %s)\n"+
		// 	"if err != nil {\n"+
		// 	"return common.NewResult(c, err)\n"+
		// 	"}\n"+
		// 	"id, err := c.ParamsInt(\"id\")\n"+
		// 	"if err != nil {\n"+
		// 	"return common.NewResult(c, err)\n"+
		// 	"}\n"+
		// 	"ctx := %s.AuthEnt.GetContext(c)\n"+
		// 	"_, err = %s.%s.UpdateByID(ctx, %s, id)\n"+
		// 	"return common.OkOrNewResult(c, err)",
		// 	paraName, entityName, paraName, receiverName, receiverName, serviceName, paraName))

		AddBody(fmt.Sprintf(""+
			"id := c.Locals(\"id\").(int)\n"+
			"ctx := c.Locals(\"ctx\").(context.Context)\n"+
			"%s, err := %s.%s.QueryByID(ctx, id)\n"+
			"if err != nil {\n"+
			"return common.NewResult(c, err)\n"+
			"}\n"+
			"err = common.BodyParser(c, %s)\n"+
			"if err != nil {\n"+
			"return common.NewResult(c, err)\n"+
			"}\n"+
			"data, err := %s.%s.UpdateByID(ctx, %s, id)\n"+
			"return common.NewResult(c, err, data)",
			paraName, receiverName, serviceName, paraName, receiverName, serviceName, paraName))

	// 根据ID删除
	group.AddLineComment(
		fmt.Sprintf(common.SwaggerCommentPrefix,
			common.DeleteByID+common.DeleteByIDComment+entityName,
			common.DeleteByID+common.DeleteByIDComment+entityName,
			common.DeleteByID+common.DeleteByIDComment+entityName,
			entityName,
			fmt.Sprintf("@Param id path int true \"%s ID\"\n", entityName),
			"common.Message",
			fmt.Sprintf("/%s/%s/%s/{id} [delete]", common.ApiName, common.ApiVersion, routerName),
		))
	deleteFun := group.NewFunction(common.DeleteByID).
		// 接收
		WithReceiver(receiverName, "*"+controllerName).
		// 参数
		AddParameter("c", "*fiber.Ctx").
		// 返回
		AddResult("", "error")
	// 方法体
	deleteFun.AddBody(fmt.Sprintf(""+
		"id := c.Locals(\"id\").(int)\n"+
		"ctx := c.Locals(\"ctx\").(context.Context)\n"+
		"err := %s.%s.DeleteByID(ctx, id)\n"+
		"return common.NewResult(c, err)",
		receiverName, serviceName))

	// 根据IDs 批量删除
	group.AddLineComment(
		fmt.Sprintf(common.SwaggerCommentPrefix,
			common.DeleteBulk+common.DeleteBulkComment+entityName,
			common.DeleteBulk+common.DeleteBulkComment+entityName,
			common.DeleteBulk+common.DeleteBulkComment+entityName,
			entityName,
			"@Param ids body common.DeleteItem true \"需要删除的id列表\"",
			"common.Message",
			fmt.Sprintf("/%s/%s/%s/bulk/delete [post]", common.ApiName, common.ApiVersion, routerName),
		))
	deleteBulkFun := group.NewFunction(common.DeleteBulk).
		// 接收
		WithReceiver(receiverName, "*"+controllerName).
		// 参数
		AddParameter("c", "*fiber.Ctx").
		// 返回
		AddResult("", "error")
	// 方法体
	deleteBulkFun.AddBody(fmt.Sprintf(""+
		"deleteItem := &common.DeleteItem{}\n"+
		"err := common.RequestBodyParser(c, &deleteItem)\n"+
		"if err != nil {\n"+
		"return common.NewResult(c, err)\n"+
		"}\n"+
		"ctx := c.Locals(\"ctx\").(context.Context)\n"+
		"_, err = %s.%s.DeleteBulk(ctx, deleteItem.Ids)\n"+
		"return common.NewResult(c, err)",
		receiverName, serviceName))

	// User 用户登录、退出
	if entityName == "User" {
		groupImport.AddPath("github.com/one-meta/meta/pkg/jwt").
			AddPath("github.com/one-meta/meta/app/entity/config").
			AddPath("github.com/one-meta/meta/app/ent/tenant").
			AddPath("github.com/one-meta/meta/app/entity").
			AddPath("errors").
			AddPath("github.com/casbin/casbin/v2").
			AddPath("github.com/one-meta/meta/pkg/auth").
			AddPath("github.com/redis/go-redis/v9")
		groupStruct.
			AddField("Enf", "*casbin.Enforcer").
			AddField("AuthEnt", "*auth.Entx").
			AddField("Rdb", "*redis.Client")
		// 登录方法
		group.AddLineComment(
			fmt.Sprintf(common.SwaggerCommentPrefix,
				common.Login+common.LoginComment+entityName,
				common.Login+common.LoginComment+entityName,
				common.Login+common.LoginComment+entityName,
				entityName,
				fmt.Sprintf("@Param %s body ent.%s true \"%s\"", pkgName, entityName, entityName),
				"common.Result{data=entity.UserInfo}",
				fmt.Sprintf("/%s/%s/%s/login [post]", common.ApiName, common.ApiVersion, routerName),
			))

		group.NewFunction(common.Login).
			// 接收
			WithReceiver(receiverName, "*"+controllerName).
			// 参数
			AddParameter("c", "*fiber.Ctx").
			// 返回
			AddResult("", "error").
			// 方法体
			AddBody("u := &ent.User{}\n" +
				"err := common.BodyParser(c, u)\n" +
				"if err != nil && (u.Name == \"\" || u.Password ==\"\") {\n" +
				"return common.NewResult(c, err)\n" +
				"}\n" +
				"ctx := context.Background()\n" +
				"u, token, err := uc.UserService.Login(ctx, u)\n" +
				"if err != nil {\n" +
				"return common.NewResult(c, err)\n" +
				"}\n" +
				"if !u.Valid {\n" +
				"return common.NewErrorWithStatusCode(c, \"User invalid\", fiber.StatusForbidden)\n" +
				"}\n" +
				"loginInfo := &entity.LoginInfo{\n" +
				"Token: token,\n" +
				"}\n" +
				"var projects []entity.Project\n" +
				"userDomains, _ := uc.Enf.GetDomainsForUser(u.Name)\n" +
				"saContext := uc.AuthEnt.GetSAContext()\n" +
				"for _, v := range userDomains {\n" +
				"queryProject, err := uc.UserService.Dao.Tenant.Query().Where(tenant.Code(v)).Only(saContext)\n" +
				"if err != nil {\n" +
				"uc.Logger.Sugar().Info(err)\n" +
				"}\n" +
				"project := entity.Project{\n" +
				"Name: queryProject.Name,\n" +
				"Code: v,\n" +
				"}\n" +
				"projects = append(projects, project)\n" +
				"}\n" +
				"if len(projects) > 0 {\n" +
				"loginInfo.Projects = projects\n" +
				"}\n" +
				"return common.NewResult(c, err, loginInfo)")

		// 退出
		group.AddLineComment(
			fmt.Sprintf(common.SwaggerCommentPrefix,
				common.Logout+common.LogoutComment+entityName,
				common.Logout+common.LogoutComment+entityName,
				common.Logout+common.LogoutComment+entityName,
				entityName,
				fmt.Sprintf("@Param %s body ent.%s true \"%s\"", pkgName, entityName, entityName),
				"common.Result{data=entity.UserInfo}",
				fmt.Sprintf("/%s/%s/%s/logout [get]", common.ApiName, common.ApiVersion, routerName),
			))
		group.NewFunction(common.Logout).
			// 接收
			WithReceiver(receiverName, "*"+controllerName).
			// 参数
			AddParameter("c", "*fiber.Ctx").
			// 返回
			AddResult("", "error").
			// 方法体
			AddBody(fmt.Sprintf("" +
				"authorization := c.Get(\"Authorization\")\n" +
				"_, err := jwt.ParseJWT(authorization, config.CFG.Auth.JWT.Key)\n" +
				"if err != nil {\n" +
				"	return common.NewResult(c, err)\n" +
				"}\n" +
				"return common.NewResult(c, err)"))

		//获取用户信息
		group.AddLineComment(
			fmt.Sprintf(common.SwaggerCommentPrefix,
				common.UserInfo+common.UserInfoComment+entityName,
				common.UserInfo+common.UserInfoComment+entityName,
				common.UserInfo+common.UserInfoComment+entityName,
				entityName,
				"",
				"common.Message",
				fmt.Sprintf("/%s/%s/%s/info [get]", common.ApiName, common.ApiVersion, routerName),
			))
		group.NewFunction(common.UserInfo).
			// 接收
			WithReceiver(receiverName, "*"+controllerName).
			// 参数
			AddParameter("c", "*fiber.Ctx").
			// 返回
			AddResult("", "error").
			// 方法体
			AddBody("if c.Locals(\"casbinUser\") == nil {\n" +
				"return common.NewResult(c, errors.New(\"nil casbin user\"))\n" +
				"}\n" +
				"authorization := c.Get(\"Authorization\")\n" +
				"if authorization == \"\" {\n" +
				"if c.Locals(\"Authorization\") != nil {\n" +
				"authorization = c.Locals(\"Authorization\").(string)\n" +
				"} else {\n" +
				"return common.NewErrorWithStatusCode(c, \"Access denied\", fiber.StatusForbidden)\n" +
				"}}\n" +
				"casbinUser := c.Locals(\"casbinUser\").(*entity.CasbinUser)\n" +
				"ctx := c.Locals(\"ctx\").(context.Context)\n" +
				"userName := casbinUser.UserName\n" +
				"project := casbinUser.Project\n" +
				"queryUser, err := uc.UserService.QueryByUserName(ctx, userName)\n" +
				"if err != nil {\n" +
				"return common.NewResult(c, err)\n" +
				"}\n" +
				"userInfo := &entity.UserInfo{}\n" +
				"if queryUser.Valid {\n" +
				"admin := queryUser.SuperAdmin\n" +
				"if admin {\n" +
				"userInfo.Role = \"admin\"\n" +
				"userAccess := &entity.Access{\n" +
				"Query:      true,\n" +
				"New:        true,\n" +
				"Edit:       true,\n" +
				"ViewDetail: true,\n" +
				"View:       true,\n" +
				"Delete:     true,\n" +
				"BulkDelete: true,\n" +
				"}\n" +
				"userInfo.Access = *userAccess\n" +
				"} else {\n" +
				"userInfo.Role = \"user\"\n" +
				"forUser, err := uc.Enf.GetRolesForUser(userName, project)\n" +
				"if err != nil {\n" +
				"return common.NewResult(c, err)\n" +
				"}\n" +
				"userAccess := &entity.Access{}\n" +
				"for _, v := range forUser {\n" +
				"switch v {\n" +
				"case \"query\":\n" +
				"userAccess.Query = true\n" +
				"case \"new\":\n" +
				"userAccess.New = true\n" +
				"case \"edit\":\n" +
				"userAccess.Edit = true\n" +
				"case \"viewDetail\":\n" +
				"userAccess.ViewDetail = true\n" +
				"case \"view\":\n" +
				"userAccess.View = true\n" +
				"case \"delete\":\n" +
				"userAccess.Delete = true\n" +
				"case \"bulkDelete\":\n" +
				"userAccess.BulkDelete = true\n" +
				"}}\n" +
				"if userAccess != nil {\n" +
				"userInfo.Access = *userAccess\n" +
				"}}\n" +
				"userInfo.Name = userName\n" +
				"loginInfo := &entity.LoginInfo{\n" +
				"Token: authorization,\n" +
				"}\n" +
				"var projects []entity.Project\n" +
				"userDomains, _ := uc.Enf.GetDomainsForUser(queryUser.Name)\n" +
				"saContext := uc.AuthEnt.GetSAContext()\n" +
				"for _, v := range userDomains {\n" +
				"queryProject, err := uc.UserService.Dao.Tenant.Query().Where(tenant.Code(v)).Only(saContext)\n" +
				"if err != nil {\n" +
				"uc.Logger.Sugar().Info(err)\n" +
				"}\n" +
				"project := entity.Project{\n" +
				"Name: queryProject.Name,\n" +
				"Code: v,\n" +
				"}\n" +
				"projects = append(projects, project)\n" +
				"}\n" +
				"if len(projects) > 0 {\n" +
				"loginInfo.Projects = projects\n" +
				"userInfo.LoginInfo = *loginInfo\n" +
				"}}\n" +
				"return common.NewResult(c, err, userInfo)")
	}
	// fmt.Println(group.String())
	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, util.Snake(entityName), generator)
}
