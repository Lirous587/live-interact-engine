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
