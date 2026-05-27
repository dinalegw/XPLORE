import { useEffect, useRef } from "react";
import { useChatStore } from "../../store/chatStore";

function formatTime(ts) {
  if (!ts) return "";
  return new Date(ts).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

export default function MessageList({ session, messages, send }) {
  const { activeRoom, typingUsers, readReceipts, onlineUsers } = useChatStore();
  const bottomRef = useRef(null);
  const myId = session.user.id;
  const typing = typingUsers[activeRoom?.id] || {};

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    const last = messages[messages.length - 1];
    if (last && last.sender_id !== myId) {
      send("receipt.read", { message_id: last.id, room_id: activeRoom?.id });
    }
  }, [messages.length]);

  return (
    <div className="message-list">
      {messages.length === 0 && (
        <div className="messages-empty">
          <span>🌱</span>
          <p>No messages yet. Start the conversation!</p>
        </div>
      )}

      {messages.map((msg) => {
        const isMine = msg.sender_id === myId;
        const readers = readReceipts[msg.id] || [];
        const senderName = msg.sender?.username || (isMine ? "You" : "Unknown");
        const senderInitial = senderName[0]?.toUpperCase() || "?";

        return (
          <div key={msg.id} className={`message-row ${isMine ? "mine" : "theirs"}`}>
            {!isMine && (
              <div className="avatar-col">
                <span className="msg-sender-name">{senderName}</span>
                <div className="avatar-wrap">
                  <div className="avatar">{senderInitial}</div>
                  <div className={`presence-dot ${onlineUsers.has(msg.sender_id) ? "online" : "offline"}`} />
                </div>
              </div>
            )}

            <div className="message-bubble-wrap">
              <div className={`message-bubble ${isMine ? "bubble-mine" : "bubble-theirs"}`}>
                {msg.content && <p className="msg-text">{msg.content}</p>}
                {msg.file_url && msg.file_url !== "" && (
                  msg.file_url.match(/\.(jpg|jpeg|png|gif|webp)$/i)
                    ? <img src={msg.file_url} className="msg-image" alt="attachment" />
                    : <a href={msg.file_url} target="_blank" className="msg-file" rel="noreferrer">📎 Attachment</a>
                )}
                <span className="msg-time">{formatTime(msg.created_at)}</span>
              </div>
              {isMine && readers.length > 0 && <span className="read-receipt">✓✓ Read</span>}
            </div>
          </div>
        );
      })}

      {Object.keys(typing).length > 0 && (
        <div className="typing-indicator">
          <div className="typing-dots"><span /><span /><span /></div>
          <span className="typing-text">
            {Object.values(typing).join(", ")} {Object.keys(typing).length === 1 ? "is" : "are"} typing...
          </span>
        </div>
      )}

      <div ref={bottomRef} />
    </div>
  );
}