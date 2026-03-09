package handler

// 专用于处理http请求

import (
	"nft_backend/internal/app/web/response"
	"nft_backend/internal/logger"
	"nft_backend/internal/person"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PersonHandler struct {
	svc *person.Service
}

// 构造函数
func NewPersonHandler(svc *person.Service) *PersonHandler {
	return &PersonHandler{svc: svc}
}

// 路由处理方法
func (h *PersonHandler) Create(c *gin.Context) {
	userID := c.GetInt("userId") // 从请求头中获取用户ID
	// 创建一个person记录
	var pf person.PersonForm
	if err := c.ShouldBindJSON(&pf); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	p := person.Person{
		Name:   pf.Name,
		Age:    pf.Age,
		City:   pf.City,
		Sex:    pf.Sex,
		UserID: userID,
	}
	h.svc.CreatePerson(&p)
	response.OK(c, "", p)
}

func (h *PersonHandler) Delete(c *gin.Context) {
	idstr := c.Param("id")
	ownerId := ownerIdFromContext(c)
	//role := c.GetString("role")    // 从请求头中获取角色
	id, err := strconv.Atoi(idstr) // 字符串转int
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	p, err := h.svc.GetPerson(id, ownerId)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	if err := h.svc.DeletePerson(id, ownerId); err != nil {
		response.Fail(c, 500, "删除person记录失败")
		return
	}
	response.OK(c, "", p)
}

func (h *PersonHandler) Update(c *gin.Context) {
	var person person.Person
	if err := c.ShouldBindJSON(&person); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 查询是否存在
	ownerId := ownerIdFromContext(c)
	id := person.ID
	_, err := h.svc.GetPerson(id, ownerId)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}

	// 修改记录
	if err := h.svc.UpdatePerson(&person, ownerId); err != nil {
		response.Fail(c, 500, "更新person记录失败")
		return
	}
	response.OK(c, "", person)
}

func (h *PersonHandler) Get(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr) // 字符串转int
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	ownerId := ownerIdFromContext(c)
	data, err := h.svc.GetPerson(id, ownerId)
	if err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	response.OK(c, "", data)
}

func (h *PersonHandler) PageList(c *gin.Context) {
	userId := c.GetInt("userId")
	logger.L.Info("分页查询用户id: " + strconv.Itoa(userId))
	var pq person.PersonPageQuery
	logger.L.Info("分页查询...")
	if err := c.ShouldBindQuery(&pq); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	pageNo := pq.PageNo
	pageSize := pq.PageSize
	if pageNo < 1 {
		pq.PageNo = 1
	}
	if pageSize < 1 {
		pq.PageSize = 10
	}
	ownerId := ownerIdFromContext(c)
	page, err := h.svc.GetPersonPage(pq, ownerId)
	if err != nil {
		response.Fail(c, 500, "分页查询失败")
		return
	}
	response.OK(c, "", page)
}

func ownerIdFromContext(c *gin.Context) *int {
	if c.GetString("role") == "admin" {
		return nil
	}
	uid := c.GetInt("userId")
	return &uid
}
