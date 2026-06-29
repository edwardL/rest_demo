package workflow

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"rest_demo/pkg/llm"
)

type MessageStore interface {
	Create(msg *SessionMessage) (int64, error)
	Update(id int64, content string) error
}

type NotifyWriter struct {
	store     MessageStore
	sessionID int64
	chatID    int64
	agentType int

	mu             sync.Mutex
	thinkID        int64
	contentID      int64
	hiddenID       int64
	contentBuf     strings.Builder
	contentTag     string
	viewTag        string
	viewPushed     bool
	lastMsgType    llm.MessageType
	suppressNotify bool

	onSSE func(event string, msg *SSEMessage)
}

func NewNotifyWriter(store MessageStore, sessionID, chatID int64, onSSE func(event string, msg *SSEMessage)) *NotifyWriter {
	return &NotifyWriter{
		store:     store,
		sessionID: sessionID,
		chatID:    chatID,
		onSSE:     onSSE,
	}
}

func (nw *NotifyWriter) SetAgentType(t int) { nw.agentType = t }

func (nw *NotifyWriter) SetViewTag(tag string) {
	nw.mu.Lock()
	defer nw.mu.Unlock()
	nw.viewTag = tag
}

func (nw *NotifyWriter) Write(full string, chunk string, msgType llm.MessageType, files []FileItem) error {
	nw.mu.Lock()
	defer nw.mu.Unlock()

	var targetID *int64
	var msgTypeDB int8
	switch msgType {
	case llm.MessageTypeThink:
		targetID = &nw.thinkID
		msgTypeDB = MsgTypeThink
	case llm.MessageTypeContent:
		msgTypeDB = MsgTypeContent
		targetID = &nw.contentID

		if nw.lastMsgType == llm.MessageTypeThink {
			if nw.thinkID > 0 {
				nw.pushSSE(nw.thinkID, nw.chatID, "</think>", false, llm.MessageTypeThink, nil)
				nw.pushSSE(nw.thinkID, nw.chatID, "", true, llm.MessageTypeThink, nil)
				nw.thinkID = 0
			}
			nw.contentID = 0
			nw.contentBuf.Reset()

			if nw.viewTag != "" && !nw.viewPushed {
				nw.viewTag = strings.TrimSpace(nw.viewTag)
				now := time.Now().Format("2006-01-02 15:04:05")
				msg := &SessionMessage{
					SessionID: nw.sessionID, ChatID: nw.chatID, AgentType: nw.agentType,
					Role: RoleAssistant, Type: MsgTypeContent, Content: nw.viewTag,
					CreateTime: now, UpdateTime: now,
				}
				if id, err := nw.store.Create(msg); err == nil {
					*targetID = id
				} else {
					*targetID = 0
				}
				nw.pushSSE(*targetID, nw.chatID, nw.viewTag, false, llm.MessageTypeContent, nil)
				nw.viewPushed = true
				nw.contentID = 0
				nw.contentBuf.Reset()
			}
		}
	}

	data := chunk
	if targetID != nil && *targetID == 0 {
		if msgType == llm.MessageTypeThink {
			data = "<think>" + chunk
		}
	}

	if *targetID == 0 {
		now := time.Now().Format("2006-01-02 15:04:05")
		msg := &SessionMessage{
			SessionID: nw.sessionID, ChatID: nw.chatID, AgentType: nw.agentType,
			Role: RoleAssistant, Type: msgTypeDB, Content: full,
			CreateTime: now, UpdateTime: now,
		}
		id, err := nw.store.Create(msg)
		if err != nil {
			return fmt.Errorf("写入消息失败: %w", err)
		}
		*targetID = id
		if msgType == llm.MessageTypeContent {
			nw.contentBuf.WriteString(full)
		}
	} else {
		if msgType == llm.MessageTypeContent {
			nw.contentBuf.WriteString(chunk)
			_ = nw.store.Update(*targetID, nw.contentBuf.String())
		} else {
			_ = nw.store.Update(*targetID, full)
		}
	}

	nw.pushSSE(*targetID, nw.chatID, data, false, msgType, files)
	nw.lastMsgType = msgType
	return nil
}

func (nw *NotifyWriter) WriteLLM(full string, chunk string, msgType llm.MessageType) error {
	return nw.Write(full, chunk, msgType, nil)
}

func (nw *NotifyWriter) NotifyFunc() NotifyFunc {
	return func(full, chunk string, msgType llm.MessageType, files []FileItem) error {
		return nw.Write(full, chunk, msgType, files)
	}
}

func (nw *NotifyWriter) Close() {
	nw.closeThink()

	nw.mu.Lock()
	needContentDone := nw.contentID > 0
	nw.mu.Unlock()

	if needContentDone {
		nw.mu.Lock()
		if nw.contentID > 0 {
			nw.pushSSE(nw.contentID, nw.chatID, "", true, llm.MessageTypeContent, nil)
		}
		nw.mu.Unlock()
	}
}

// CloseThink 关闭当前 think 块，推送 </think> + done。
func (nw *NotifyWriter) CloseThink() {
	nw.closeThink()
}

func (nw *NotifyWriter) closeThink() {
	nw.mu.Lock()
	defer nw.mu.Unlock()
	if nw.thinkID > 0 {
		nw.pushSSE(nw.thinkID, nw.chatID, "</think>", false, llm.MessageTypeThink, nil)
		nw.pushSSE(nw.thinkID, nw.chatID, "", true, llm.MessageTypeThink, nil)
		nw.thinkID = 0
	}
	nw.viewPushed = false
}

func (nw *NotifyWriter) pushSSE(id, chatID int64, data string, done bool, msgType llm.MessageType, files []FileItem) {
	if nw.suppressNotify {
		return
	}
	if nw.onSSE == nil {
		return
	}
	if files == nil {
		files = make([]FileItem, 0)
	}
	msg := &SSEMessage{
		Event:       SSEEventMessage,
		Message:     data,
		MessageType: string(msgType),
		Role:        RoleAssistant,
		ID:          id,
		Done:        done,
		SessionID:   nw.sessionID,
		AgentType:   nw.agentType,
		ChatID:      chatID,
		Files:       files,
	}
	nw.onSSE(SSEEventMessage, msg)
}
