package person

type Person struct {
	ID     int    `json:"id" gorm:"primaryKey"` // 主键, 注意: 字段名必须为ID, 且字段类型必须为int, gorm:"primaryKey"
	Name   string `json:"name"`
	Age    int    `json:"age"`
	City   string `json:"city"`
	Sex    string `json:"sex"`
	UserID int    `json:"user_id"`
}

type PersonPageEntity struct {
	PageNo   int      `json:"pageNo"`
	PageSize int      `json:"pageSize"`
	Total    int      `json:"total"`
	List     []Person `json:"list"`
}

type PersonPageQuery struct {
	PageNo   int    `form:"pageNo"  min:"1" default:"1"`
	PageSize int    `form:"pageSize" min:"1" default:"10"`
	City     string `form:"city"`
	Sex      string `form:"sex"`
}

// 定义一个json接收结构体, 用于接收前端提交的json数据
type PersonForm struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age" binding:"required"`
	City string `json:"city" binding:"required"`
	Sex  string `json:"sex" binding:"required"`
}
