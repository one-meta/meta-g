package common

import (
	"fmt"
	"os"

	mapset "github.com/deckarep/golang-set"
)

var (
	Separator         = string(os.PathSeparator)
	Query             = "Query"
	QueryComment      = " 根据指定字段、时间范围查询或搜索 "
	QueryByID         = "QueryByID"
	QueryByIDComment  = " 根据 ID 查询 "
	Create            = "Create"
	CreateComment     = " 创建 "
	CreateBulk        = "CreateBulk"
	CreateBulkComment = " 批量创建 "
	UpdateByID        = "UpdateByID"
	UpdateByIDComment = " 根据 ID 修改 "
	DeleteByID        = "DeleteByID"
	DeleteByIDComment = " 根据 ID 删除 "
	DeleteBulk        = "DeleteBulk"
	DeleteBulkComment = " 根据 IDs 批量删除 "

	// user
	Login           = "Login"
	LoginComment    = " 根据用户名和密码登录 "
	Logout          = "Logout"
	LogoutComment   = " 根据退出登录 "
	UserInfo        = "UserInfo"
	UserInfoComment = " 获取用户信息 "

	QueryPage          = "QueryPage"
	QueryPageComment   = " 分页查询 "
	QuerySearch        = "QuerySearch"
	QuerySearchComment = " 分页搜索 "

	ApiPath        = "api"
	ControllerPath = "app"
	ServicePath    = "app"
	ExtendPath     = fmt.Sprintf("app%sent", Separator)

	ApiName    = "api"
	ApiVersion = "v1"
	ApiTest    = "test"

	SwaggerCommentPrefix = `
    %s
    @Description %s
    @Summary %s
    @Tags %s
    @Accept json
    @Produce json
    %s
    @Success 200 {object} %s
    @Router %s
`

	DataMap        map[string]string
	DefaultDataMap map[string][]string
	TimeSet        mapset.Set
	FieldMap       map[string][]StructField

	Merge bool
)

type StructField struct {
	FiledName   string `json:"filedName"`
	FieldLength string `json:"fieldLength"`
	FieldType   string `json:"fieldType"`
	Nillable    bool   `json:"nillable"`
}
