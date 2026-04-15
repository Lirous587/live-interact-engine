package handler

import (
	"live-interact-engine/services/api-service/internal/adapter/mapper"
	"live-interact-engine/services/api-service/internal/utils/ctxutil"
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

	ownerID, ok := ctxutil.GetUserID(ctx)
	if !ok || ownerID == "" {
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

	// 同样改掉这里
	ownerID, ok := ctxutil.GetUserID(ctx)
	if !ok || ownerID == "" {
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

// RemoveRole 移除用户权限 API
// @Summary 移除用户权限
// @Description 房间所有者移除用户权限
// @Tags Room
// @Accept json
// @Produce json
// @Param request body mapper.RemoveRoleReq true "移除权限请求"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /v1/room/remove-role [post]
func (h *RoomHandler) RemoveRole(ctx *gin.Context) {
	var req mapper.RemoveRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 获取房间所有者ID
	ownerID, ok := ctxutil.GetUserID(ctx)
	if !ok || ownerID == "" {
		response.Error(ctx, errors.New("unauthorized: user_id not found"))
		return
	}

	// 调用 mapper
	err := h.roomMapper.RemoveRole(ctx.Request.Context(), ownerID, &req)
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
// @Param permission path int true "权限值"
// @Success 200 {object} mapper.CheckPermissionResp "权限检查结果"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /v1/room/{room_id}/user/{user_id}/permission/{permission} [get]
func (h *RoomHandler) CheckPermission(ctx *gin.Context) {

	// 从 query 参数获取 permission
	var req mapper.CheckPermissionReq
	if err := ctx.ShouldBindUri(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 调用 mapper
	resp, err := h.roomMapper.CheckPermission(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// MuteUser 禁言用户 API
// @Summary 禁言用户
// @Description 房间管理员禁言指定用户
// @Tags Room
// @Accept json
// @Produce json
// @Param request body mapper.MuteUserReq true "禁言用户请求"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /v1/room/mute [post]
func (h *RoomHandler) MuteUser(ctx *gin.Context) {
	var req mapper.MuteUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	adminID, ok := ctxutil.GetUserID(ctx)
	if !ok || adminID == "" {
		response.Error(ctx, errors.New("unauthorized: user_id not found"))
		return
	}

	err := h.roomMapper.MuteUser(ctx.Request.Context(), adminID, &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, gin.H{})
}

// UnmuteUser 解除禁言 API
// @Summary 解除禁言
// @Description 房间管理员解除用户禁言
// @Tags Room
// @Accept json
// @Produce json
// @Param request body mapper.UnmuteUserReq true "解除禁言请求"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Router /v1/room/unmute [post]
func (h *RoomHandler) UnmuteUser(ctx *gin.Context) {
	var req mapper.UnmuteUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	adminID, ok := ctxutil.GetUserID(ctx)
	if !ok || adminID == "" {
		response.Error(ctx, errors.New("unauthorized: user_id not found"))
		return
	}

	req.SetAdminID(adminID)

	err := h.roomMapper.UnmuteUser(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, gin.H{})
}

// IsMuted 检查用户禁言状态 API
// @Summary 检查禁言状态
// @Description 检查用户是否在房间中被禁言
// @Tags Room
// @Produce json
// @Param room_id path string true "房间ID"
// @Param user_id path string true "用户ID"
// @Success 200 {object} mapper.IsMutedResp "禁言状态"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /v1/room/{room_id}/user/{user_id}/mute-status [get]
func (h *RoomHandler) IsMuted(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	userID := ctx.Param("user_id")

	if roomID == "" || userID == "" {
		response.InvalidParams(ctx, errors.New("room_id and user_id required"))
		return
	}

	req := &mapper.IsMutedReq{
		RoomID: roomID,
		UserID: userID,
	}

	resp, err := h.roomMapper.IsMuted(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// GetMuteInfo 获取禁言信息 API
// @Summary 获取禁言信息
// @Description 获取用户在房间的禁言详细信息
// @Tags Room
// @Produce json
// @Param room_id path string true "房间ID"
// @Param user_id path string true "用户ID"
// @Success 200 {object} mapper.MuteResp "禁言信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /v1/room/{room_id}/user/{user_id}/mute-info [get]
func (h *RoomHandler) GetMuteInfo(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	userID := ctx.Param("user_id")

	if roomID == "" || userID == "" {
		response.InvalidParams(ctx, errors.New("room_id and user_id required"))
		return
	}

	req := &mapper.GetMuteInfoReq{
		RoomID: roomID,
		UserID: userID,
	}

	resp, err := h.roomMapper.GetMuteInfo(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}

// GetMuteList 获取禁言列表 API
// @Summary 获取禁言列表
// @Description 获取房间中被禁言的用户列表
// @Tags Room
// @Produce json
// @Param room_id path string true "房间ID"
// @Param offset query int false "分页偏移量" default(0)
// @Param limit query int false "分页大小" default(10)
// @Success 200 {object} mapper.GetMuteListResp "禁言列表"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Router /v1/room/{room_id}/mute-list [get]
func (h *RoomHandler) GetMuteList(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	if roomID == "" {
		response.InvalidParams(ctx, errors.New("room_id required"))
		return
	}

	var req mapper.GetMuteListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 10
	}

	req.RoomID = roomID

	resp, err := h.roomMapper.GetMuteList(ctx.Request.Context(), &req)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
