package apicodes

// 弹幕模块错误码 (1200-1399)
var (
	ErrDanmakuNeedRoomID = ErrCode{Msg: "room_id required", Type: ErrorTypeBadRequest, Code: 1200}
)
