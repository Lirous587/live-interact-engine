package handler

import (
	"live-interact-engine/services/api-service/internal/adapter/mapper"
	"live-interact-engine/services/api-service/internal/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type RoomHandler struct {
	roomMapper *mapper.RoomMapper
}

func NewRoomHandler(roomMapper *mapper.RoomMapper) *RoomHandler {
	return &RoomHandler{
		roomMapper: roomMapper,
	}
}

// CreateRoom 创建房间 API
// @Summary 创建房间
// @Description 用户创建一个新房间
// @Tags Room
// @Accept json
// @Produce json
// @Param request body mapper.CreateRoomReq true "创建房间请求"
// @Success 200 {object} mapper.RoomResp "房间信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /v1/room/create [post]
func (h *RoomHandler) CreateRoom(ctx *gin.Context) {
	var req mapper.CreateRoomReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 从认证信息或请求头获取 owner_id（需要认证中间件）
	ownerID := ctx.GetString("user_id")
	if ownerID == "" {
		response.Error(ctx, errors.New("unauthorized: user_id not found"))
		return
	}
	req.OwnerID = ownerID

	// 调用 mapper
	resp, err := h.roomMapper.CreateRoom(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// GetRoom 获取房间信息 API
// @Summary 获取房间信息
// @Description 根据房间ID获取房间详细信息
// @Tags Room
// @Produce json
// @Param room_id path string true "房间ID"
// @Success 200 {object} mapper.RoomResp "房间信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 404 {object} map[string]interface{} "房间不存在"
// @Router /v1/room/{room_id} [get]
func (h *RoomHandler) GetRoom(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	if roomID == "" {
		response.InvalidParams(ctx, errors.New("room_id required"))
		return
	}

	req := &mapper.GetRoomReq{RoomID: roomID}

	// 调用 mapper
	resp, err := h.roomMapper.GetRoom(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// AssignRole 分配用户角色 API
// @Summary 分配用户角色
// @Description 房间所有者为用户分配角色和权限
// @Tags Room
// @Accept json
// @Produce json
// @Param request body mapper.AssignRoleReq true "分配角色请求"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /v1/room/assign-role [post]
func (h *RoomHandler) AssignRole(ctx *gin.Context) {
	var req mapper.AssignRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 从认证信息获取 owner_id（只有房主可以分配角色）
	ownerID := ctx.GetString("user_id")
	if ownerID == "" {
		response.Error(ctx, errors.New("unauthorized: user_id not found"))
		return
	}

	// 调用 mapper
	err := h.roomMapper.AssignRole(ctx.Request.Context(), ownerID, &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, gin.H{})
}

// GetUserRoomRole 获取用户在房间的角色 API
// @Summary 获取用户角色
// @Description 获取用户在指定房间的角色和权限信息
// @Tags Room
// @Produce json
// @Param room_id path string true "房间ID"
// @Param user_id path string true "用户ID"
// @Success 200 {object} mapper.UserRoomRoleResp "用户角色信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 404 {object} map[string]interface{} "角色不存在"
// @Router /v1/room/{room_id}/user/{user_id}/role [get]
func (h *RoomHandler) GetUserRoomRole(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	userID := ctx.Param("user_id")

	if roomID == "" || userID == "" {
		response.InvalidParams(ctx, errors.New("room_id and user_id required"))
		return
	}

	req := &mapper.GetUserRoomRoleReq{
		RoomID: roomID,
		UserID: userID,
	}

	// 调用 mapper
	resp, err := h.roomMapper.GetUserRoomRole(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// CheckPermission 检查用户权限 API
// @Summary 检查用户权限
// @Description 检查用户是否拥有指定权限
// @Tags Room
// @Produce json
// @Param room_id path string true "房间ID"
// @Param user_id path string true "用户ID"
// @Param permission query int true "权限值"
// @Success 200 {object} mapper.CheckPermissionResp "权限检查结果"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /v1/room/{room_id}/user/{user_id}/permission [get]
func (h *RoomHandler) CheckPermission(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	userID := ctx.Param("user_id")

	if roomID == "" || userID == "" {
		response.InvalidParams(ctx, errors.New("room_id and user_id required"))
		return
	}

	// 从 query 参数获取 permission
	var req mapper.CheckPermissionReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	req.RoomID = roomID
	req.UserID = userID

	// 调用 mapper
	resp, err := h.roomMapper.CheckPermission(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
