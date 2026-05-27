export default function RoomItem({ room, active, onClick }) {
  return (
    <button className={`room-item ${active ? "room-item--active" : ""}`} onClick={onClick}>
      <span className="room-icon">{room.is_private ? "🔒" : "🌿"}</span>
      <span className="room-name">{room.name}</span>
    </button>
  );
}