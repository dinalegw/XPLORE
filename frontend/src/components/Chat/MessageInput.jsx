import { useState, useRef } from "react";
import { useChatStore } from "../../store/chatStore";
import api from "../../lib/api";

export default function MessageInput({ session, send }) {
  const { activeRoom, currentUser } = useChatStore();
  const [text, setText] = useState("");
  const [uploading, setUploading] = useState(false);
  const fileRef = useRef(null);
  const typingTimer = useRef(null);
  const isTyping = useRef(false);

  // Use username from currentUser profile, fallback to email
  const username = currentUser?.username || session.user.email;

  const handleTyping = (val) => {
    setText(val);
    if (!isTyping.current) {
      isTyping.current = true;
      send("typing.start", { room_id: activeRoom.id, user_id: session.user.id, username: username });
    }
    clearTimeout(typingTimer.current);
    typingTimer.current = setTimeout(() => {
      isTyping.current = false;
      send("typing.stop", { room_id: activeRoom.id, user_id: session.user.id });
    }, 1500);
  };

  const handleSend = () => {
    if (!text.trim()) return;
    send("message.send", { room_id: activeRoom.id, content: text.trim(), file_url: "" });
    setText("");
    isTyping.current = false;
    clearTimeout(typingTimer.current);
    send("typing.stop", { room_id: activeRoom.id, user_id: session.user.id });
  };

  const handleFile = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    setUploading(true);
    const form = new FormData();
    form.append("file", file);
    try {
      const res = await api.post("/api/upload", form, { headers: { "Content-Type": "multipart/form-data" } });
      send("message.send", { room_id: activeRoom.id, content: "", file_url: res.data.url });
    } catch (err) {
      console.error("Upload failed", err);
    }
    setUploading(false);
    e.target.value = "";
  };

  return (
    <div className="message-input-bar glass-panel-dark">
      <input ref={fileRef} type="file" style={{ display: "none" }} onChange={handleFile} accept="image/*,.pdf,.doc,.docx,.txt" />
      <button className="attach-btn" onClick={() => fileRef.current.click()} disabled={uploading}>{uploading ? "⏳" : "📎"}</button>
      <input className="text-input glass-input" type="text" placeholder="Say something..." value={text} onChange={(e) => handleTyping(e.target.value)} onKeyDown={(e) => e.key === "Enter" && handleSend()} />
      <button className="send-btn btn-wood" onClick={handleSend} disabled={!text.trim()}>Send</button>
    </div>
  );
}