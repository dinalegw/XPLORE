import { useEffect, useRef, useCallback } from "react";
import { useChatStore } from "../store/chatStore";

export function useWebSocket(session, activeRoom) {
  const wsRef = useRef(null);
  const { addMessage, setTyping, setUserOnline, setUserOffline, addReadReceipt } = useChatStore();

  const send = useCallback((type, payload) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }));
    } else {
      console.warn("WebSocket not open, state:", wsRef.current?.readyState);
    }
  }, []);

  useEffect(() => {
    if (!session || !activeRoom) return;

    // Close existing connection first
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    const userId = session.user.id;
    const roomId = activeRoom.id;
    const wsUrl = `${import.meta.env.VITE_WS_URL || "ws://localhost:8080"}/ws?user_id=${userId}&room_id=${roomId}`;

    console.log("Connecting WebSocket to", wsUrl);
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => console.log("WebSocket connected for room", roomId);

    ws.onmessage = (e) => {
      const event = JSON.parse(e.data);
      const p = event.payload;
      console.log("WS event received:", event.type, p);
      switch (event.type) {
        case "message.new":
          addMessage(p.room_id, p);
          break;
        case "typing.start":
          setTyping(p.room_id, p.user_id, p.username, true);
          break;
        case "typing.stop":
          setTyping(p.room_id, p.user_id, p.username, false);
          break;
        case "presence.online":
          setUserOnline(p.user_id);
          break;
        case "presence.offline":
          setUserOffline(p.user_id);
          break;
        case "receipt.read":
          addReadReceipt(p.message_id, p.user_id);
          break;
      }
    };

    ws.onerror = (e) => console.error("WebSocket error", e);
    ws.onclose = () => console.log("WebSocket closed");

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [session?.user?.id, activeRoom?.id]);

  return { send };
}