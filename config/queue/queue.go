package queue

// TODO: 消息队列配置

const (
	/**	TODO: 交换机 */

	// EventExchange 系统事件交换机（用于广播所有服务器）
	EventExchange = "system_event_"

	/**	TODO: 队列	*/

	// QDelayEmailTask 后台延时发送邮件队列
	QDelayEmailTask = "DELAY_EMAIL_TASK"

	/**	TODO: 事件	*/

	// CloseServer 停服
	CloseServer = "STOP_SERVER"
	// RepeatLogin 重复登录
	RepeatLogin = "REPEAT_LOGIN"
	// NewEmail 新邮件
	NewEmail = "NEW_EMAIL"
	// GetAttachment 邮件领取附件
	GetAttachment = "GET_ATTACHMENT"
)

// Event 消息结构体
type Event struct {
	Type          string      `json:"type"`           // 事件类型
	Val           interface{} `json:"val"`            // 事件值（可以在该字段存放任意数据，复杂类型使用json）
	ExcludeServer []string    `json:"exclude_server"` // 不需要执行该事件的server
	GameId        int32       `json:"game_id"`        // 标识具体游戏
	Range         Range       `json:"range"`
}

// UIdRId 定义的序列号结构体
type UIdRId struct {
	UId    int64 `json:"u_id"`
	RoleId int64 `json:"role_id"`
}

// Range 定义的区间
type Range struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
