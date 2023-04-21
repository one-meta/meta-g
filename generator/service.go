package generator

import (
	"fmt"
	"github.com/one-meta/meta-g/common"
	"github.com/one-meta/meta-g/util"
	"strings"

	"github.com/Xuanwo/gg"
)

// GenService 生成Service代码
func GenService(entityName, rootPath string) {
	daoName := entityName + "Service"
	receiverName := util.Receiver(daoName)
	paraName := util.Receiver(entityName)
	queryParamName := util.Receiver("queryParam")
	lowerEntityName := strings.ToLower(entityName)

	generator := gg.New()
	group := generator.NewGroup()

	packageName := "service"
	// 包名
	group.AddPackage(packageName)

	// import
	path := group.NewImport().
		AddPath("context").
		AddPath("github.com/one-meta/meta/app/ent").
		AddPath("github.com/one-meta/meta/app/entity")

	path.AddPath("github.com/one-meta/meta/app/ent/" + lowerEntityName)

	structField := group.NewStruct(daoName).AddField("Dao", "*Dao")

	// 查询 搜索
	group.AddLineComment(common.Query + common.QueryComment + entityName)
	group.NewFunction(common.Query).
		// 接收
		WithReceiver(receiverName, "*"+daoName).
		// 参数
		AddParameter("ctx", "context.Context").AddParameter(paraName, "*ent."+entityName).AddParameter(queryParamName, "*entity.QueryParam").
		// 返回
		AddResult("", "int").AddResult("", "[]*ent."+entityName).AddResult("", "error").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"if len(%s.Search) == 0 {\n"+
			"return %s.QueryPage(ctx, %s, %s)\n"+
			"} else {\n"+
			"return %s.QuerySearch(ctx, %s, %s)}",
			queryParamName, receiverName, paraName, queryParamName, receiverName, paraName, queryParamName))

	// 根据ID查询
	group.AddLineComment(common.QueryByID + common.QueryByIDComment + entityName)
	group.NewFunction(common.QueryByID).
		// 接收
		WithReceiver(receiverName, "*"+daoName).
		// 参数
		AddParameter("ctx", "context.Context").AddParameter("id", "int").
		// 返回
		AddResult("", "*ent."+entityName).AddResult("", "error").
		// 方法体
		// AddBody(fmt.Sprintf("return %s.Dao.%s.Get(ctx, id)", receiverName, entityName))
		AddBody(fmt.Sprintf("return %s.Dao.%s.Query().Where(%s.ID(id)).Only(ctx)", receiverName, entityName, lowerEntityName))

	// 创建
	group.AddLineComment(common.Create + common.CreateComment + entityName)
	createFun := group.NewFunction(common.Create).
		// 接收
		WithReceiver(receiverName, "*"+daoName).
		// 参数
		AddParameter("ctx", "context.Context").AddParameter(paraName, "*ent."+entityName).
		// 返回
		AddResult("", "*ent."+entityName).AddResult("", "error")
	// 方法体
	if entityName != "User" {
		if entityName == "CasbinRule" {
			createFun.AddBody(fmt.Sprintf(""+
				"var err error\n"+
				"switch cr.Type {\n"+
				"case \"g\":\n"+
				"_, err = %s.Enf.AddGroupingPolicy(cr.Sub, cr.Dom, cr.Obj)\n"+
				"case \"p\":\n"+
				"_, err = %s.Enf.AddPolicy(cr.Sub, cr.Dom, cr.Obj, cr.Act)\n"+
				"}\n"+
				"if err != nil {\n"+
				"return nil, err\n"+
				"}\n"+
				"return cr,err",
				receiverName, receiverName))
		} else {
			createFun.AddBody(fmt.Sprintf("return %s.Dao.%s.Create().SetItem%s(%s).Save(ctx)", receiverName, entityName, entityName, paraName))
		}
	}

	defaultValues := common.DefaultDataMap[entityName]
	var defaultValue string
	if len(defaultValues) != 0 {
		var defaultValue2 []string
		for _, v := range defaultValues {
			key := strings.Split(v, ":")[0]
			value := strings.Split(v, ":")[1]
			defaultValue2 = append(defaultValue2, "v."+key+"="+value)
		}
		defaultValue = strings.Join(defaultValue2, "\n")
	}

	// 批量创建
	group.AddLineComment(common.CreateBulk + common.CreateBulkComment + entityName)
	createBulkFun := group.NewFunction(common.CreateBulk).
		// 接收
		WithReceiver(receiverName, "*"+daoName).
		// 参数
		AddParameter("ctx", "context.Context").AddParameter(paraName, "[]*ent."+entityName).
		// 返回
		AddResult("", "[]*ent."+entityName).AddResult("", "error")
	// 方法体
	if entityName != "User" {
		if entityName == "CasbinRule" {
			createBulkFun.AddBody(fmt.Sprintf(""+
				"var (\n"+
				"err         error\n"+
				"policyRules [][]string\n"+
				"groupRules  [][]string\n"+
				")\n"+
				"for _, rule := range cr {\n"+
				"err = common.CheckCasbinRule(rule)\n"+
				"if err != nil {\n"+
				"return []*ent.CasbinRule{rule}, err\n"+
				"}\n"+
				"switch rule.Type {\n"+
				"case \"g\":\n"+
				"groupRules = append(groupRules, []string{rule.Sub, rule.Dom, rule.Obj})\n"+
				"case \"p\":\n"+
				"policyRules = append(policyRules, []string{rule.Sub, rule.Dom, rule.Obj, rule.Act})\n"+
				"}\n"+
				"}\n"+
				"if len(policyRules) != 0 {\n"+
				"_, err = %s.Enf.AddPolicies(policyRules)\n"+
				"}\n"+
				"if len(groupRules) != 0 {\n"+
				"_, err = %s.Enf.AddGroupingPolicies(groupRules)\n"+
				"}\n"+
				"if err != nil {\n"+
				"return nil, err\n"+
				"}\nreturn cr, nil",
				receiverName, receiverName))
		} else {
			if len(defaultValues) != 0 {
				createBulkFun.AddBody(fmt.Sprintf(""+
					"bulks := make([]*ent.%sCreate, len(%s))\n"+
					"for i, v := range %s {\n"+
					"%s\n"+
					"bulks[i] = %s.Dao.%s.Create().SetItem%s(v)\n"+
					"}\n"+
					"return %s.Dao.%s.CreateBulk(bulks...).Save(ctx)",
					entityName, paraName, paraName, defaultValue, receiverName, entityName, entityName, receiverName, entityName))
			} else {
				createBulkFun.AddBody(fmt.Sprintf(""+
					"bulks := make([]*ent.%sCreate, len(%s))\n"+
					"for i, v := range %s {\n"+
					"bulks[i] = %s.Dao.%s.Create().SetItem%s(v)\n"+
					"}\n"+
					"return %s.Dao.%s.CreateBulk(bulks...).Save(ctx)",
					entityName, paraName, paraName, receiverName, entityName, entityName, receiverName, entityName))
			}
		}
	}

	// 修改
	group.AddLineComment(common.UpdateByID + common.UpdateByIDComment + entityName)
	updateFun := group.NewFunction(common.UpdateByID).
		// 接收
		WithReceiver(receiverName, "*"+daoName).
		// 参数
		AddParameter("ctx", "context.Context").AddParameter(paraName, "*ent."+entityName).AddParameter("id", "int").
		// 返回
		AddResult("", "*ent."+entityName).AddResult("", "error")
	// 方法体
	if entityName != "User" {
		if entityName == "CasbinRule" {
			updateFun.AddBody(fmt.Sprintf(""+
				"var err error\n"+
				"dbRule, err := %s.QueryByID(ctx, id)\n"+
				"crStr := getRuleSlice(cr)\n"+
				"dbStr := getRuleSlice(dbRule)\n"+
				"if cr.Type == dbRule.Type {\n"+
				"switch cr.Type {\n"+
				"case \"g\":\n"+
				"_, err = %s.Enf.UpdateGroupingPolicy(dbStr, crStr)\n"+
				"case \"p\":\n"+
				"_, err = %s.Enf.UpdatePolicy(dbStr, crStr)\n"+
				"}\n"+
				"}"+
				" else {\n"+
				"return nil, errors.New(\"dismactch rule type\")\n"+
				"}\nif err != nil {\n"+
				"return nil, err\n}\n"+
				"return cr, nil",
				receiverName, receiverName, receiverName))
		} else {
			updateFun.AddBody(fmt.Sprintf(
				"return %s.Dao.%s.UpdateOneID(id).SetItem%s(%s).Save(ctx)",
				receiverName, entityName, entityName, paraName))
		}
	}

	// 根据ID删除
	group.AddLineComment(common.DeleteByID + common.DeleteByIDComment + entityName)
	deleteFun := group.NewFunction(common.DeleteByID).
		// 接收
		WithReceiver(receiverName, "*"+daoName).

		// 返回
		AddResult("", "error")
	deleteFun.
		// 参数
		AddParameter("ctx", "context.Context").AddParameter("id", "int")
	// 方法体
	if entityName == "CasbinRule" {
		deleteFun.
			AddBody(fmt.Sprintf(""+
				"rule, err := %s.QueryByID(ctx, id)\n"+
				"if err != nil {\n"+
				"return err\n"+
				"}\n"+
				"switch rule.Type {\n"+
				"case \"g\":\n"+
				"_, err = %s.Enf.RemoveGroupingPolicy(rule.Sub, rule.Dom, rule.Obj)\n"+
				"case \"p\":\n"+
				"_, err = %s.Enf.RemovePolicy(rule.Sub, rule.Dom, rule.Obj, rule.Act)\n"+
				"}\n"+
				"if err != nil {\n"+
				"return err\n"+
				"}\n"+
				"return nil",
				receiverName, receiverName, receiverName))
	} else {
		deleteFun.
			AddBody(fmt.Sprintf("return %s.Dao.%s.DeleteOneID(id).Exec(ctx)", receiverName, entityName))
	}

	// 根据IDs 批量删除
	group.AddLineComment(common.DeleteBulk + common.DeleteBulkComment + entityName)
	deleteBulkFun := group.NewFunction(common.DeleteBulk).
		// 接收
		WithReceiver(receiverName, "*"+daoName).
		// 返回
		AddResult("", "int").
		AddResult("", "error").
		// 参数
		AddParameter("ctx", "context.Context").AddParameter("ids", "[]int")
	// 方法体
	if entityName == "CasbinRule" {
		deleteBulkFun.
			AddBody(fmt.Sprintf(""+
				"for _, id := range ids {\n"+
				"err := %s.DeleteByID(ctx, id)\n"+
				"if err != nil {\n"+
				"return 1, err\n"+
				"}\n"+
				"}\n"+
				"return 0, nil", receiverName))
	} else {
		deleteBulkFun.
			AddBody(fmt.Sprintf(""+
				"count, err := %s.Dao.%s.Delete().Where(%s.IDIn(ids...)).Exec(ctx)\n",
				receiverName, entityName, lowerEntityName) +
				"return count ,err",
			)
	}

	// 分页查询
	group.AddLineComment(common.QueryPage + common.QueryPageComment + entityName)
	group.NewFunction(common.QueryPage).
		// 接收
		WithReceiver(receiverName, "*"+daoName).
		// 参数
		AddParameter("ctx", "context.Context").AddParameter(paraName, "*ent."+entityName).AddParameter(queryParamName, "*entity.QueryParam").
		// 返回
		AddResult("", "int").AddResult("", "[]*ent."+entityName).AddResult("", "error").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"count, err := %s.Dao.%s.Query().QueryItem%s(%s, %s, true).Count(ctx)\n"+
			"if err != nil {\n"+
			"return 0, nil, err\n"+
			"}\n"+
			"results, err := %s.Dao.%s.Query().QueryItem%s(%s, %s, false).All(ctx)\n"+
			"if err != nil {\n"+
			"return 0, nil, err\n"+
			"}\n"+
			"if len(results) == 0 {\n"+
			"count = 0\n"+
			"}\n"+
			"return count, results, nil",
			receiverName, entityName, entityName, paraName, queryParamName, receiverName, entityName, entityName, paraName, queryParamName))

	// 分页搜索
	group.AddLineComment(common.QuerySearch + common.QuerySearchComment + entityName)
	group.NewFunction(common.QuerySearch).
		// 接收
		WithReceiver(receiverName, "*"+daoName).
		// 参数
		AddParameter("ctx", "context.Context").AddParameter(paraName, "*ent."+entityName).AddParameter(queryParamName, "*entity.QueryParam").
		// 返回
		AddResult("", "int").AddResult("", "[]*ent."+entityName).AddResult("", "error").
		// 方法体
		AddBody(fmt.Sprintf(""+
			"count, err := %s.Dao.%s.Query().Search%s(%s, %s, true).Count(ctx)\n"+
			"if err != nil {\n"+
			"return 0, nil, err\n"+
			"}\n"+
			"results, err := %s.Dao.%s.Query().Search%s(%s, %s, false).All(ctx)\n"+
			"if err != nil {\n"+
			"return 0, nil, err\n"+
			"}\n"+
			"if len(results) == 0 {\n"+
			"count = 0\n"+
			"}\n"+
			"return count, results, nil",
			receiverName, entityName, entityName, paraName, queryParamName, receiverName, entityName, entityName, paraName, queryParamName))

	// User 用户登录
	if entityName == "User" {
		// import
		path.
			AddPath("github.com/one-meta/meta/app/ent/user").
			AddPath("github.com/one-meta/meta/pkg/bcrypt").
			AddPath("github.com/one-meta/meta/pkg/jwt").
			AddPath("github.com/one-meta/meta/app/entity/config").
			AddPath("errors")

		// create function
		createFun.AddBody(fmt.Sprintf(""+
			"encodePassword, err := bcrypt.Encode(u.Password)\n"+
			"if err != nil {\n"+
			"return nil, err\n"+
			"}\n"+
			"u.Password = encodePassword\n"+
			"return %s.Dao.%s.Create().SetItem%s(%s).Save(ctx)", receiverName, entityName, entityName, paraName))

		// create bulk function
		createBulkFun.AddBody(fmt.Sprintf(""+"for _,v := range u{\n"+
			"encodePassword, err := bcrypt.Encode(v.Password)\n"+
			"if err != nil {\n"+
			"return nil, err\n"+
			"}\n"+
			"%s\n"+
			"v.Password = encodePassword\n"+
			"}\n"+
			"bulks := make([]*ent.%sCreate, len(%s))\n"+
			"for i, v := range %s {\n"+
			"bulks[i] = %s.Dao.%s.Create().SetItem%s(v)\n"+
			"}\n"+
			"return %s.Dao.%s.CreateBulk(bulks...).Save(ctx)",
			defaultValue, entityName, paraName, paraName, receiverName, entityName, entityName, receiverName, entityName))

		// update function
		updateFun.AddBody(fmt.Sprintf(""+
			"encodePassword, err := bcrypt.Encode(u.Password)\n"+
			"if err != nil {\n"+
			"return nil, err\n"+
			"}\n"+
			"u.Password = encodePassword\n"+
			"return %s.Dao.%s.UpdateOneID(id).SetItem%s(%s).Save(ctx)\n",
			receiverName, entityName, entityName, paraName))

		// login function
		group.AddLineComment(common.Login + common.LoginComment + entityName)
		group.NewFunction(common.Login).
			// 接收
			WithReceiver(receiverName, "*"+daoName).
			// 参数
			AddParameter("ctx", "context.Context").AddParameter(paraName, "*ent."+entityName).
			// 返回
			AddResult("", "*ent."+entityName).AddResult("", "string").AddResult("", "error").
			// 方法体
			AddBody(fmt.Sprintf(""+
				"%s, err := %s.Dao.%s.Query().Where(user.Name(%s.Name)).First(ctx)\n"+
				"if err != nil {\n"+
				"return nil, \"\", errors.New(\"login error\")\n"+
				"}\n"+
				"matches := bcrypt.Matches(%s.Password, %s.Password)\n"+
				"if matches {\n"+
				"jwtConfig := config.CFG.Auth.JWT\n"+
				"token, err := jwt.CreateJWT(\"\", %s.Name, jwtConfig.Key, jwtConfig.TTL, jwtConfig.Hmac)\n"+
				"if err != nil {\n"+
				"return nil, \"\", err\n"+
				"}\n"+
				"return %s,token, nil\n"+
				"}\n"+
				"return nil,\"\", errors.New(\"login error\")",
				queryParamName, receiverName, entityName, paraName, queryParamName, paraName, queryParamName, queryParamName))

		// 根据用户名查询用户
		group.AddLineComment("QueryByUserName 根据 Name 查询 User")
		group.NewFunction("QueryByUserName").
			// 接收
			WithReceiver(receiverName, "*"+daoName).
			// 参数
			AddParameter("ctx", "context.Context").AddParameter("name", "string").
			// 返回
			AddResult("", "*ent."+entityName).AddResult("", "error").
			// 方法体
			AddBody("return us.Dao.User.Query().Where(user.Name(name)).Only(ctx)")

	}

	// 额外数据
	if entityName == "CasbinRule" {
		path.AddPath("github.com/one-meta/meta/pkg/common").
			AddPath("errors").
			AddPath("github.com/casbin/casbin/v2")
		structField.AddField("Enf", "*casbin.Enforcer")
		// getRuleSlice
		group.AddLineComment("getRuleSlice")
		group.NewFunction("getRuleSlice").
			// 参数
			AddParameter("rule", "*ent."+entityName).
			// 返回
			AddResult("", "[]string").
			// 方法体
			AddBody(
				"switch rule.Type {\n" +
					"case \"g\":\n" +
					"return []string{rule.Sub, rule.Dom, rule.Obj}\n" +
					"case \"p\":\n" +
					"return []string{rule.Sub, rule.Dom, rule.Obj, rule.Act}\n" +
					"default:\n" +
					"return nil\n" +
					"}")
	}
	// fmt.Println(group.String())
	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, util.Snake(entityName), generator)
}

// GenServiceWireSet 生成Service wire set
func GenServiceWireSet(rootPath string) {
	generator := gg.New()
	group := generator.NewGroup()
	packageName := "service"
	// 包名
	group.AddPackage(packageName)

	// import
	group.NewImport().AddPath("github.com/google/wire").AddPath("github.com/one-meta/meta/app/ent")

	group.AddLineComment("Dao 所有的Dao都是ent Client")
	group.AddTypeAlias("Dao", "ent.Client")
	setGroup := make([]string, 0, len(common.DataMap))
	for k := range common.DataMap {
		setGroup = append(setGroup, k+"Set")
	}
	newVar := group.NewVar().AddField("\nSet", fmt.Sprintf("wire.NewSet(%s)", strings.Join(setGroup, ",")))
	for _, v := range setGroup {
		serviceSet := strings.ReplaceAll(v, "Set", "Service")
		newVar.AddField(v, fmt.Sprintf("wire.NewSet(wire.Struct(new(%s), \"*\"))", serviceSet))
	}

	// fmt.Println(group.String())
	util.CheckDirAndMk(rootPath, packageName)
	util.CheckGenFile(rootPath, packageName, "wire_set", generator)
}
