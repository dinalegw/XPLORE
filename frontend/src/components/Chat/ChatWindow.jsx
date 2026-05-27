import { useEffect } from "react";
import { useChatStore } from "../../store/chatStore";
import MessageList from "./MessageList";
import MessageInput from "./MessageInput";
import api from "../../lib/api";

export default function ChatWindow({ session, send }) {
  const { activeRoom, messages, setMessages } = useChatStore();
  const roomMessages = messages[activeRoom?.id] || [];

  useEffect(() => {
    if (!activeRoom) return;
    api.get(`/api/rooms/${activeRoom.id}/messages`).then((res) => {
      setMessages(activeRoom.id, res.data || []);
    });
  }, [activeRoom?.id]);

  return (
    <div className="chat-window glass-panel">
      <div className="chat-header">
        <div className="chat-header-info">
          <span className="chat-header-icon">{activeRoom?.is_private ? "🔒" : "🌿"}</span>
          <div>
            <h2 className="chat-header-name">{activeRoom?.name}</h2>
            <p className="chat-header-type">{activeRoom?.is_private ? "Private Chat" : "Group Chat"}</p>
          </div>
        </div>
      </div>
      <MessageList session={session} messages={roomMessages} send={send} />
      <MessageInput session={session} send={send} />
    </div>
  );
}