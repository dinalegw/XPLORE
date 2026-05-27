import { useState, useEffect, useRef } from "react";
import api from "../../lib/api";

export default function UserSearch({ onClose, onRequestSent }) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState([]);
  const [friendRequests, setFriendRequests] = useState([]);
  const [tab, setTab] = useState("search");
  const [loading, setLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const inputRef = useRef(null);

  useEffect(() => {
    inputRef.current?.focus();
    fetchFriendRequests();
  }, []);

  const fetchFriendRequests = async () => {
    const res = await api.get("/api/users/friend-requests");
    setFriendRequests(res.data || []);
  };

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      setHasSearched(false);
      return;
    }
    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await api.get(`/api/users/search?q=${query}`);
        setResults(res.data || []);
        setHasSearched(true);
      } catch (err) { console.error(err); }
      setLoading(false);
    }, 400);
    return () => clearTimeout(timer);
  }, [query]);

  const handleSendRequest = async (userId) => {
    try {
      await api.post("/api/users/friend-request", { receiver_id: userId });
      setResults((prev) =>
        prev.map((u) => u.id === userId ? { ...u, friend_status: "pending" } : u)
      );
      // Notify parent to refresh friends list
      if (onRequestSent) {
        onRequestSent();
      }
    } catch (err) { console.error(err); }
  };

  const handleRespond = async (requestId, approve) => {
    try {
      const res = await api.post(`/api/users/friend-requests/${requestId}/respond`, { approve });
      setFriendRequests((prev) => prev.filter((r) => r.id !== requestId));
      if (approve && res.data.room_id) {
        const roomsRes = await api.get("/api/rooms");
        const { setRooms } = (await import("../../store/chatStore")).useChatStore.getState();
        setRooms(roomsRes.data || []);
        onClose();
      }
    } catch (err) { console.error(err); }
  };

  const getFriendButton = (user) => {
    if (user.friend_status === "pending") {
      return <span style={{ fontSize: "12px", color: "var(--text-muted)" }}>⏳ Pending</span>;
    }
    if (user.friend_status === "accepted") {
      return <span style={{ fontSize: "12px", color: "var(--accent)" }}>✓ Friends</span>;
    }
    return (
      <button className="btn-wood" style={{ padding: "6px 14px", fontSize: "13px" }}
        onClick={() => handleSendRequest(user.id)}>
        + Add
      </button>
    );
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal glass-panel user-search-modal" onClick={(e) => e.stopPropagation()}>
        <h3 className="modal-title">👤 Find People</h3>

        <div className="auth-tabs" style={{ marginBottom: "16px" }}>
          <button className={`auth-tab ${tab === "search" ? "active" : ""}`} onClick={() => setTab("search")}>
            Search
          </button>
          <button className={`auth-tab ${tab === "requests" ? "active" : ""}`} onClick={() => setTab("requests")}>
            Requests {friendRequests.length > 0 && `(${friendRequests.length})`}
          </button>
        </div>

        {tab === "search" && (
          <>
            <input
              ref={inputRef}
              className="glass-input"
              placeholder="Search by username..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
            <div className="browse-list" style={{ marginTop: "12px" }}>
              {loading && (
                <p style={{ color: "var(--text-muted)", fontSize: "14px", textAlign: "center", padding: "10px" }}>
                  Searching...
                </p>
              )}
              {!loading && hasSearched && results.length === 0 && (
                <p style={{ color: "var(--text-muted)", fontSize: "14px", textAlign: "center", padding: "10px" }}>
                  No users found
                </p>
              )}
              {!loading && results.map((user) => (
                <div key={user.id} className="browse-item">
                  <div className="avatar-wrap">
                    <div className="avatar avatar-sm">{user.username[0].toUpperCase()}</div>
                    <div className={`presence-dot ${user.is_online ? "online" : "offline"}`} />
                  </div>
                  <div style={{ flex: 1 }}>
                    <p style={{ color: "var(--text-primary)", fontSize: "15px" }}>{user.username}</p>
                    <p style={{ color: "var(--text-muted)", fontSize: "12px" }}>
                      {user.is_online ? "🟢 Online" : "⚫ Offline"}
                    </p>
                  </div>
                  {getFriendButton(user)}
                </div>
              ))}
            </div>
          </>
        )}

        {tab === "requests" && (
          <div className="browse-list">
            {friendRequests.length === 0 && (
              <p style={{ color: "var(--text-muted)", fontSize: "14px", textAlign: "center", padding: "20px 0" }}>
                No pending friend requests
              </p>
            )}
            {friendRequests.map((req) => (
              <div key={req.id} className="browse-item">
                <div className="avatar avatar-sm">{req.username[0].toUpperCase()}</div>
                <div style={{ flex: 1 }}>
                  <p style={{ color: "var(--text-primary)", fontSize: "14px" }}>{req.username}</p>
                  <p style={{ color: "var(--text-muted)", fontSize: "12px" }}>wants to be friends</p>
                </div>
                <div style={{ display: "flex", gap: "6px" }}>
                  <button className="btn-wood" style={{ padding: "5px 12px", fontSize: "12px" }}
                    onClick={() => handleRespond(req.id, true)}>✓</button>
                  <button className="btn-ghost" style={{ padding: "5px 12px", fontSize: "12px" }}
                    onClick={() => handleRespond(req.id, false)}>✗</button>
                </div>
              </div>
            ))}
          </div>
        )}

        <div className="modal-actions" style={{ marginTop: "16px" }}>
          <button className="btn-ghost" onClick={onClose}>Close</button>
        </div>
      </div>
    </div>
  );
}