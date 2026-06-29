package workflow

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

const (
	MsgTypeContent  int8 = 0
	MsgTypeThink    int8 = 1
	MsgTypeToolCall int8 = 2
	MsgTypeFile     int8 = 3
	MsgTypeHidden   int8 = 9
)

const (
	SSEEventMessage       = "message"
	SSEEventUpdate        = "update"
	SSEEventUpdateSession = "update_session"
)

type FileItem struct {
	FileName string `json:"file_name"`
	FileID   string `json:"file_id"`
	FileSize int64  `json:"file_size,omitempty"`
}

type SessionMessage struct {
	ID         int64  `json:"id"`
	SessionID  int64  `json:"session_id"`
	ChatID     int64  `json:"chat_id"`
	AgentType  int    `json:"agent_type"`
	Role       string `json:"role"`
	Type       int8   `json:"type"`
	Content    string `json:"content"`
	ExtInfo    string `json:"ext_info,omitempty"`
	Status     int    `json:"status"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

type SSEMessage struct {
	Event       string     `json:"event"`
	Message     string     `json:"message"`
	MessageType string     `json:"message_type,omitempty"`
	Role        string     `json:"role,omitempty"`
	ID          int64      `json:"id,omitempty"`
	Done        bool       `json:"done,omitempty"`
	SessionID   int64      `json:"session_id"`
	AgentType   int        `json:"agent_type"`
	ChatID      int64      `json:"chat_id"`
	Files       []FileItem `json:"files,omitempty"`
}
